[CmdletBinding()]
param(
    [switch]$DryRun,
    [switch]$Resume,
    [switch]$PushSource,
    [switch]$BootstrapPullSecret,
    [ValidatePattern('^[0-9a-f]{40}$')]
    [string]$RollbackTag = '',
    [ValidatePattern('^ccr\.ccs\.tencentyun\.com/[a-z0-9._-]+/[a-z0-9._-]+$')]
    [string]$ImageBase = 'ccr.ccs.tencentyun.com/littlehorse/personal-page-be',
    [ValidatePattern('^[A-Za-z0-9._-]+@[A-Za-z0-9._:-]+$')]
    [string]$SshTarget = 'ubuntu@124.223.181.152',
    [ValidatePattern('^[a-z0-9]([-a-z0-9]*[a-z0-9])?$')]
    [string]$Namespace = 'default',
    [ValidatePattern('^[a-z0-9]([-a-z0-9]*[a-z0-9])?$')]
    [string]$Deployment = 'personal-page-be-deployment',
    [ValidatePattern('^[a-z0-9]([-a-z0-9]*[a-z0-9])?$')]
    [string]$Container = 'personal-page-be-container',
    [ValidatePattern('^[a-z0-9]([-a-z0-9]*[a-z0-9])?$')]
    [string]$PullSecret = 'ccr-tencent',
    [ValidateRange(60, 900)]
    [int]$TimeoutSeconds = 240
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$RepoRoot = Split-Path -Parent $PSScriptRoot
$Registry = 'ccr.ccs.tencentyun.com'
$ExpectedBranch = 'master'
$ExpectedKubeContext = 'kubernetes-admin@kubernetes'
$PodSelector = 'app=personal-page-be-pod'
$PublicHealthUrl = 'https://api.peacesheep.xyz/api/ping'
$SourceUrl = 'https://github.com/li1553770945/personal-page-be'
if ($Resume -and $RollbackTag) {
    throw '-Resume and -RollbackTag cannot be used together.'
}
if ($PushSource -and $RollbackTag) {
    throw '-PushSource cannot be used with -RollbackTag.'
}

function Invoke-NativeCapture {
    param(
        [Parameter(Mandatory = $true)][string]$FilePath,
        [Parameter(Mandatory = $true)][string[]]$Arguments,
        [string]$WorkingDirectory = $RepoRoot
    )

    Push-Location $WorkingDirectory
    $previousErrorActionPreference = $ErrorActionPreference
    try {
        $ErrorActionPreference = 'Continue'
        $output = & $FilePath @Arguments 2>&1
        $exitCode = $LASTEXITCODE
    }
    finally {
        $ErrorActionPreference = $previousErrorActionPreference
        Pop-Location
    }

    $text = ($output | Out-String).Trim()
    if ($exitCode -ne 0) {
        throw "$FilePath failed with exit code ${exitCode}:`n$text"
    }
    return $text
}

function Invoke-Native {
    param(
        [Parameter(Mandatory = $true)][string]$FilePath,
        [Parameter(Mandatory = $true)][string[]]$Arguments,
        [string]$WorkingDirectory = $RepoRoot
    )

    Push-Location $WorkingDirectory
    $previousErrorActionPreference = $ErrorActionPreference
    try {
        $ErrorActionPreference = 'Continue'
        & $FilePath @Arguments
        $exitCode = $LASTEXITCODE
    }
    finally {
        $ErrorActionPreference = $previousErrorActionPreference
        Pop-Location
    }

    if ($exitCode -ne 0) {
        throw "$FilePath failed with exit code $exitCode."
    }
}

function Invoke-NativeResult {
    param(
        [Parameter(Mandatory = $true)][string]$FilePath,
        [Parameter(Mandatory = $true)][string[]]$Arguments,
        [string]$WorkingDirectory = $RepoRoot
    )

    Push-Location $WorkingDirectory
    $previousErrorActionPreference = $ErrorActionPreference
    try {
        $ErrorActionPreference = 'Continue'
        $output = & $FilePath @Arguments 2>&1
        $exitCode = $LASTEXITCODE
    }
    finally {
        $ErrorActionPreference = $previousErrorActionPreference
        Pop-Location
    }

    return [pscustomobject]@{
        ExitCode = $exitCode
        Output = ($output | Out-String).Trim()
    }
}

function Invoke-RemoteScript {
    param(
        [Parameter(Mandatory = $true)][string]$Script,
        [switch]$Capture
    )

    $ssh = Get-Command ssh -ErrorAction Stop
    $startInfo = New-Object System.Diagnostics.ProcessStartInfo
    $startInfo.FileName = $ssh.Source
    $startInfo.Arguments = "-o BatchMode=yes -o StrictHostKeyChecking=yes -o ConnectTimeout=10 $SshTarget `"bash -s`""
    $startInfo.UseShellExecute = $false
    $startInfo.CreateNoWindow = $true
    $startInfo.RedirectStandardInput = $true
    $startInfo.RedirectStandardOutput = $true
    $startInfo.RedirectStandardError = $true
    $process = New-Object System.Diagnostics.Process
    $process.StartInfo = $startInfo
    if (-not $process.Start()) {
        throw 'Failed to start the SSH client.'
    }
    $stdoutTask = $process.StandardOutput.ReadToEndAsync()
    $stderrTask = $process.StandardError.ReadToEndAsync()
    $process.StandardInput.Write($Script)
    $process.StandardInput.Close()
    $process.WaitForExit()
    $stdout = $stdoutTask.Result.Trim()
    $stderr = $stderrTask.Result.Trim()
    $exitCode = $process.ExitCode
    $process.Dispose()

    if ($exitCode -ne 0) {
        $details = ($stdout, $stderr | Where-Object { $_ }) -join "`n"
        throw "Remote command failed with exit code ${exitCode}:`n$details"
    }
    if ($Capture) {
        return $stdout
    }
    if ($stdout) {
        Write-Host $stdout
    }
    if ($stderr) {
        Write-Warning $stderr
    }
}

function Ensure-Docker {
    if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
        throw 'Docker CLI was not found. Install Docker Desktop first.'
    }

    $result = Invoke-NativeResult -FilePath 'docker' -Arguments @('info', '--format', '{{.ServerVersion}}')
    if ($result.ExitCode -eq 0) {
        return
    }

    $dockerDesktop = Join-Path $env:ProgramFiles 'Docker\Docker\Docker Desktop.exe'
    if (-not (Test-Path -LiteralPath $dockerDesktop)) {
        throw 'Docker daemon is not running and Docker Desktop was not found.'
    }

    Write-Host 'Docker Desktop is not running; starting it now...'
    Start-Process -FilePath $dockerDesktop -WindowStyle Hidden
    for ($attempt = 1; $attempt -le 60; $attempt++) {
        Start-Sleep -Seconds 2
        $result = Invoke-NativeResult -FilePath 'docker' -Arguments @('info', '--format', '{{.ServerVersion}}')
        if ($result.ExitCode -eq 0) {
            Write-Host 'Docker daemon is ready.'
            return
        }
    }

    throw 'Docker daemon did not become ready within 120 seconds.'
}

function Get-DockerCredential {
    $configPath = Join-Path $HOME '.docker\config.json'
    if (-not (Test-Path -LiteralPath $configPath)) {
        throw "Docker config was not found. Run: docker login $Registry"
    }

    $config = Get-Content -LiteralPath $configPath -Raw -Encoding UTF8 | ConvertFrom-Json
    $authProperty = $null
    $authsProperty = $config.PSObject.Properties | Where-Object { $_.Name -eq 'auths' } | Select-Object -First 1
    if ($authsProperty -and $authsProperty.Value) {
        $authProperty = $authsProperty.Value.PSObject.Properties | Where-Object { $_.Name -eq $Registry } | Select-Object -First 1
    }

    $storedAuthProperty = $null
    if ($authProperty -and $authProperty.Value) {
        $storedAuthProperty = $authProperty.Value.PSObject.Properties | Where-Object { $_.Name -eq 'auth' } | Select-Object -First 1
    }
    if ($storedAuthProperty -and $storedAuthProperty.Value) {
        $decoded = [Text.Encoding]::UTF8.GetString([Convert]::FromBase64String([string]$storedAuthProperty.Value))
        $separator = $decoded.IndexOf(':')
        if ($separator -lt 1) {
            throw 'The stored CCR credential is malformed.'
        }
        return [pscustomobject]@{
            Username = $decoded.Substring(0, $separator)
            Secret = $decoded.Substring($separator + 1)
        }
    }

    $helperName = ''
    $credHelpersProperty = $config.PSObject.Properties | Where-Object { $_.Name -eq 'credHelpers' } | Select-Object -First 1
    if ($credHelpersProperty -and $credHelpersProperty.Value) {
        $helperProperty = $credHelpersProperty.Value.PSObject.Properties | Where-Object { $_.Name -eq $Registry } | Select-Object -First 1
        if ($helperProperty) {
            $helperName = [string]$helperProperty.Value
        }
    }
    $credsStoreProperty = $config.PSObject.Properties | Where-Object { $_.Name -eq 'credsStore' } | Select-Object -First 1
    if (-not $helperName -and $credsStoreProperty -and $credsStoreProperty.Value) {
        $helperName = [string]$credsStoreProperty.Value
    }
    if (-not $helperName) {
        throw "No credential helper is configured for $Registry. Run: docker login $Registry"
    }

    $helper = Get-Command "docker-credential-$helperName" -ErrorAction SilentlyContinue
    if (-not $helper) {
        throw "Docker credential helper docker-credential-$helperName was not found."
    }

    $previousErrorActionPreference = $ErrorActionPreference
    try {
        $ErrorActionPreference = 'Continue'
        $rawCredential = $Registry | & $helper.Source get 2>&1
        $helperExitCode = $LASTEXITCODE
    }
    finally {
        $ErrorActionPreference = $previousErrorActionPreference
    }
    if ($helperExitCode -ne 0) {
        throw "Unable to read the stored CCR credential. Run: docker login $Registry"
    }
    $credential = ($rawCredential | Out-String) | ConvertFrom-Json
    if (-not $credential.Username -or -not $credential.Secret) {
        throw "Stored CCR credential is empty. Run: docker login $Registry"
    }

    return [pscustomobject]@{
        Username = [string]$credential.Username
        Secret = [string]$credential.Secret
    }
}

function Sync-CcrPullSecret {
    param([Parameter(Mandatory = $true)]$Credential)

    $authText = "{0}:{1}" -f $Credential.Username, $Credential.Secret
    $authValue = [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes($authText))
    $auths = [ordered]@{}
    $auths[$Registry] = [ordered]@{ auth = $authValue }
    $dockerConfig = ([ordered]@{ auths = $auths } | ConvertTo-Json -Compress -Depth 5)
    $dockerConfigBase64 = [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes($dockerConfig))

    $script = (@'
set -euo pipefail
umask 077
tmp="$(mktemp)"
trap 'rm -f "$tmp"' EXIT
printf '%s' '__DOCKER_CONFIG_BASE64__' | base64 -d > "$tmp"
kubectl -n '__NAMESPACE__' create secret generic '__PULL_SECRET__' \
  --type=kubernetes.io/dockerconfigjson \
  --from-file=.dockerconfigjson="$tmp" \
  --dry-run=client -o yaml | \
  kubectl apply --server-side --field-manager=personal-page-deploy -f -
'@).Replace('__DOCKER_CONFIG_BASE64__', $dockerConfigBase64).
    Replace('__NAMESPACE__', $Namespace).
    Replace('__PULL_SECRET__', $PullSecret)

    try {
        Invoke-RemoteScript -Script $script
    }
    finally {
        $authText = $null
        $authValue = $null
        $dockerConfig = $null
        $dockerConfigBase64 = $null
    }
}

function Get-RegistryImageInfo {
    param([Parameter(Mandatory = $true)][string]$Reference)

    $result = Invoke-NativeResult -FilePath 'docker' -Arguments @('buildx', 'imagetools', 'inspect', $Reference)
    if ($result.ExitCode -ne 0) {
        if ($result.Output -match '(?i)(manifest unknown|not found|does not exist)') {
            return [pscustomobject]@{ Exists = $false; Digest = '' }
        }
        throw "Unable to inspect $Reference. Check CCR login and network:`n$($result.Output)"
    }

    $digestMatch = [regex]::Match($result.Output, '(?m)^Digest:\s+(sha256:[0-9a-f]{64})\s*$')
    if (-not $digestMatch.Success) {
        throw "Registry returned an image but no digest could be parsed for $Reference."
    }
    return [pscustomobject]@{ Exists = $true; Digest = $digestMatch.Groups[1].Value }
}

function Get-RemoteDeploymentState {
    $script = (@'
set -euo pipefail
kubectl -n '__NAMESPACE__' get deployment '__DEPLOYMENT__' -o jsonpath='{.metadata.resourceVersion}{"|"}{.spec.template.spec.containers[0].image}'
'@).Replace('__NAMESPACE__', $Namespace).
    Replace('__DEPLOYMENT__', $Deployment)

    $rawState = (Invoke-RemoteScript -Script $script -Capture).Trim()
    $parts = $rawState -split '\|', 2
    if ($parts.Length -ne 2 -or -not $parts[0] -or -not $parts[1]) {
        throw "Unable to parse remote Deployment state: $rawState"
    }
    return [pscustomobject]@{
        ResourceVersion = $parts[0]
        Image = $parts[1]
    }
}

function Set-RemoteImage {
    param(
        [Parameter(Mandatory = $true)][string]$TargetImage,
        [Parameter(Mandatory = $true)][string]$ExpectedImage,
        [Parameter(Mandatory = $true)][string]$ExpectedResourceVersion
    )

    $script = (@'
set -euo pipefail
patch='[
  {"op":"test","path":"/metadata/resourceVersion","value":"__RESOURCE_VERSION__"},
  {"op":"test","path":"/spec/template/spec/containers/0/name","value":"__CONTAINER__"},
  {"op":"test","path":"/spec/template/spec/containers/0/image","value":"__EXPECTED_IMAGE__"},
  {"op":"replace","path":"/spec/template/spec/containers/0/image","value":"__TARGET_IMAGE__"},
  {"op":"replace","path":"/spec/template/spec/containers/0/imagePullPolicy","value":"IfNotPresent"},
  {"op":"replace","path":"/spec/template/spec/imagePullSecrets","value":[{"name":"dockerhub-regcred"},{"name":"__PULL_SECRET__"}]}
]'
kubectl -n '__NAMESPACE__' patch deployment '__DEPLOYMENT__' --type=json -p "$patch"
'@).Replace('__NAMESPACE__', $Namespace).
    Replace('__DEPLOYMENT__', $Deployment).
    Replace('__CONTAINER__', $Container).
    Replace('__PULL_SECRET__', $PullSecret).
    Replace('__RESOURCE_VERSION__', $ExpectedResourceVersion).
    Replace('__EXPECTED_IMAGE__', $ExpectedImage).
    Replace('__TARGET_IMAGE__', $TargetImage)

    Invoke-RemoteScript -Script $script
}

function Wait-RemoteRollout {
    param([Parameter(Mandatory = $true)][string]$TargetImage)

    $script = (@'
set -euo pipefail
kubectl -n '__NAMESPACE__' annotate deployment '__DEPLOYMENT__' \
  kubernetes.io/change-cause="deploy __TARGET_IMAGE__" --overwrite
kubectl -n '__NAMESPACE__' rollout status deployment/'__DEPLOYMENT__' --timeout='__TIMEOUT__s'
'@).Replace('__NAMESPACE__', $Namespace).
    Replace('__DEPLOYMENT__', $Deployment).
    Replace('__TARGET_IMAGE__', $TargetImage).
    Replace('__TIMEOUT__', [string]$TimeoutSeconds)

    Invoke-RemoteScript -Script $script
}

function Test-RemoteDeployment {
    param([Parameter(Mandatory = $true)][string]$TargetImage)

    $script = (@'
set -euo pipefail
spec_image="$(kubectl -n '__NAMESPACE__' get deployment '__DEPLOYMENT__' -o jsonpath='{.spec.template.spec.containers[0].image}')"
test "$spec_image" = '__TARGET_IMAGE__'
desired="$(kubectl -n '__NAMESPACE__' get deployment '__DEPLOYMENT__' -o jsonpath='{.spec.replicas}')"
ready="$(kubectl -n '__NAMESPACE__' get deployment '__DEPLOYMENT__' -o jsonpath='{.status.readyReplicas}')"
test "$desired" = "$ready"
for attempt in $(seq 1 30); do
  rows="$(kubectl -n '__NAMESPACE__' get pods -l '__POD_SELECTOR__' --no-headers \
    -o 'custom-columns=NAME:.metadata.name,DELETING:.metadata.deletionTimestamp,READY:.status.containerStatuses[0].ready,RESTARTS:.status.containerStatuses[0].restartCount,IMAGE:.spec.containers[0].image,IMAGE_ID:.status.containerStatuses[0].imageID')"
  active_rows="$(printf '%s\n' "$rows" | awk '$2 == "<none>" {print}')"
  active_count="$(printf '%s\n' "$active_rows" | awk 'NF {count++} END {print count+0}')"
  bad_pods="$(printf '%s\n' "$active_rows" | awk '$3 != "true" || $4 != "0" || $5 != "__TARGET_IMAGE__" {print}')"
  if [ "$active_count" = "$desired" ] && [ -z "$bad_pods" ]; then
    break
  fi
  if [ "$attempt" = '30' ]; then
    printf '%s\n' "$rows" >&2
    exit 1
  fi
  sleep 2
done
health="$(curl -fsS --retry 5 --retry-delay 2 --max-time 15 http://127.0.0.1:31902/api/ping)"
printf '%s' "$health" | grep -q '"code":0'
printf '%s\n' "$active_rows"
'@).Replace('__NAMESPACE__', $Namespace).
    Replace('__DEPLOYMENT__', $Deployment).
    Replace('__TARGET_IMAGE__', $TargetImage).
    Replace('__POD_SELECTOR__', $PodSelector)

    Invoke-RemoteScript -Script $script
}

function Restore-PreviousImage {
    param(
        [Parameter(Mandatory = $true)][string]$FailedImage,
        [Parameter(Mandatory = $true)][string]$PreviousImage
    )

    $currentState = Get-RemoteDeploymentState
    if ($currentState.Image -ne $FailedImage) {
        Write-Warning "Live image changed concurrently to $($currentState.Image); automatic rollback was skipped."
        return
    }

    Write-Warning "Rolling back explicitly to $PreviousImage"
    Set-RemoteImage -TargetImage $PreviousImage -ExpectedImage $FailedImage -ExpectedResourceVersion $currentState.ResourceVersion
    Wait-RemoteRollout -TargetImage $PreviousImage
    Test-RemoteDeployment -TargetImage $PreviousImage
}

function Assert-RemotePreflight {
    param([switch]$RequireSecretWrite)

    $secretWriteChecks = ''
    if ($RequireSecretWrite) {
        $secretWriteChecks = @'
test "$(kubectl auth can-i create secrets -n '__NAMESPACE__')" = 'yes'
test "$(kubectl auth can-i patch secrets -n '__NAMESPACE__')" = 'yes'
'@.Replace('__NAMESPACE__', $Namespace)
    }

    $script = (@'
set -euo pipefail
context="$(kubectl config current-context)"
test "$context" = '__KUBE_CONTEXT__'
kubectl -n '__NAMESPACE__' get deployment '__DEPLOYMENT__' >/dev/null
container_names="$(kubectl -n '__NAMESPACE__' get deployment '__DEPLOYMENT__' -o jsonpath='{range .spec.template.spec.containers[*]}{.name}{"\n"}{end}')"
test "$container_names" = '__CONTAINER__'
current_image="$(kubectl -n '__NAMESPACE__' get deployment '__DEPLOYMENT__' -o jsonpath='{.spec.template.spec.containers[0].image}')"
test -n "$current_image"
test "$(kubectl auth can-i patch deployments.apps -n '__NAMESPACE__')" = 'yes'
test "$(kubectl auth can-i get secrets -n '__NAMESPACE__')" = 'yes'
__SECRET_WRITE_CHECKS__
'@).Replace('__KUBE_CONTEXT__', $ExpectedKubeContext).
    Replace('__NAMESPACE__', $Namespace).
    Replace('__DEPLOYMENT__', $Deployment).
    Replace('__CONTAINER__', $Container).
    Replace('__SECRET_WRITE_CHECKS__', $secretWriteChecks)

    Invoke-RemoteScript -Script $script
}

function Assert-CcrPullSecret {
    $script = (@'
set -euo pipefail
secret_type="$(kubectl -n '__NAMESPACE__' get secret '__PULL_SECRET__' -o jsonpath='{.type}')"
test "$secret_type" = 'kubernetes.io/dockerconfigjson'
kubectl -n '__NAMESPACE__' get secret '__PULL_SECRET__' -o jsonpath='{.data.\.dockerconfigjson}' | \
  base64 -d | grep -Fq '"ccr.ccs.tencentyun.com"'
'@).Replace('__NAMESPACE__', $Namespace).
    Replace('__PULL_SECRET__', $PullSecret)

    Invoke-RemoteScript -Script $script
}

function Test-PublicHealth {
    for ($attempt = 1; $attempt -le 3; $attempt++) {
        try {
            $response = Invoke-RestMethod -Uri $PublicHealthUrl -Method Get -TimeoutSec 20
            if ([int]$response.code -eq 0) {
                return $true
            }
        }
        catch {
            Write-Warning "Public health check attempt $attempt failed: $($_.Exception.Message)"
        }
        if ($attempt -lt 3) {
            Start-Sleep -Seconds 5
        }
    }
    return $false
}

$currentSha = Invoke-NativeCapture -FilePath 'git' -Arguments @('rev-parse', 'HEAD')
$targetTag = if ($RollbackTag) { $RollbackTag } else { $currentSha }
$imageTag = "${ImageBase}:$targetTag"
$mode = if ($RollbackTag) { 'rollback' } elseif ($Resume) { 'resume' } else { 'deploy' }

Write-Host '=== personal-page-be local CCR deployment ==='
Write-Host "Mode:       $mode"
Write-Host "Image tag:  $imageTag"
Write-Host "SSH target: $SshTarget"

if ($DryRun) {
    Write-Host 'Dry run only. No tests, push, secret update, or Kubernetes change was performed.'
    return
}

if (-not $RollbackTag) {
    $branch = Invoke-NativeCapture -FilePath 'git' -Arguments @('rev-parse', '--abbrev-ref', 'HEAD')
    if ($branch -ne $ExpectedBranch) {
        throw "Deployments are only allowed from $ExpectedBranch; current branch is $branch."
    }

    $status = Invoke-NativeCapture -FilePath 'git' -Arguments @('status', '--porcelain')
    if ($status) {
        throw 'Working tree is not clean. Commit or stash changes before deployment.'
    }

    Invoke-Native -FilePath 'git' -Arguments @('fetch', 'origin', $ExpectedBranch, '--prune')
    $countsText = Invoke-NativeCapture -FilePath 'git' -Arguments @('rev-list', '--left-right', '--count', "HEAD...origin/$ExpectedBranch")
    $counts = $countsText -split '\s+'
    $localAhead = [int]$counts[0]
    $remoteAhead = [int]$counts[1]
    if ($remoteAhead -gt 0) {
        throw "origin/$ExpectedBranch contains $remoteAhead commit(s) not present locally. Pull/rebase before deployment."
    }

    if (-not $Resume) {
        Invoke-Native -FilePath 'go' -Arguments @('test', './...')
        Invoke-Native -FilePath 'go' -Arguments @('vet', './...')
    }

    if ($localAhead -gt 0) {
        if (-not $PushSource) {
            throw "Local HEAD is $localAhead commit(s) ahead of origin/$ExpectedBranch. Push it first or rerun with -PushSource."
        }
        Invoke-Native -FilePath 'git' -Arguments @('push', 'origin', "HEAD:$ExpectedBranch")
    }
    $remoteLine = Invoke-NativeCapture -FilePath 'git' -Arguments @('ls-remote', 'origin', "refs/heads/$ExpectedBranch")
    $remoteSha = ($remoteLine -split '\s+')[0]
    if ($remoteSha -ne $currentSha) {
        throw "Remote $ExpectedBranch does not match local HEAD after push."
    }
}

if (-not (Get-Command ssh -ErrorAction SilentlyContinue)) {
    throw 'OpenSSH client was not found.'
}
Assert-RemotePreflight -RequireSecretWrite:$BootstrapPullSecret
Ensure-Docker
Invoke-Native -FilePath 'docker' -Arguments @('buildx', 'version')

$imageInfo = Get-RegistryImageInfo -Reference $imageTag
if ($mode -eq 'deploy' -and $imageInfo.Exists) {
    throw "Immutable tag already exists: $imageTag. Use -Resume to deploy it; the script will never overwrite it."
}
if (($mode -eq 'resume' -or $mode -eq 'rollback') -and -not $imageInfo.Exists) {
    throw "CCR image tag does not exist: $imageTag"
}

if ($mode -eq 'deploy') {
    $builderImage = [string]$env:DOCKER_GO_BUILDER_IMAGE
    if (-not $builderImage) {
        $builderImage = 'docker.m.daocloud.io/library/golang:1.25.12-bookworm@sha256:a9c020ee3d1508c7be5435c262434e3d3fc1d0e76a11afeb9ddae7d60bc86aa4'
    }
    if ($builderImage -notmatch '@sha256:[0-9a-f]{64}$') {
        throw 'DOCKER_GO_BUILDER_IMAGE must be pinned to an explicit sha256 digest.'
    }
    Write-Host "Building with pinned Go image: $builderImage"
    $buildArguments = @(
        'buildx', 'build',
        '--platform', 'linux/amd64',
        '--provenance=false',
        '--build-arg', "GO_BUILDER_IMAGE=$builderImage",
        '--label', "org.opencontainers.image.revision=$currentSha",
        '--label', "org.opencontainers.image.source=$SourceUrl",
        '--tag', $imageTag,
        '--load', '.'
    )
    if ($env:DOCKER_BUILD_PROGRESS) {
        $buildArguments = @('buildx', 'build', '--progress', [string]$env:DOCKER_BUILD_PROGRESS) + $buildArguments[2..($buildArguments.Length - 1)]
    }
    Invoke-Native -FilePath 'docker' -Arguments $buildArguments

    $pushResult = Invoke-NativeResult -FilePath 'docker' -Arguments @('push', $imageTag)
    if ($pushResult.Output) {
        Write-Host $pushResult.Output
    }
    if ($pushResult.ExitCode -ne 0) {
        $ambiguousImageInfo = Get-RegistryImageInfo -Reference $imageTag
        if ($ambiguousImageInfo.Exists) {
            throw "Docker push returned an error after the immutable tag became visible. Rerun with -Resume; this run will not overwrite the tag."
        }
        throw "Docker push failed:`n$($pushResult.Output)"
    }

    $imageInfo = Get-RegistryImageInfo -Reference $imageTag
    if (-not $imageInfo.Exists) {
        throw "Image push completed but the CCR tag cannot be inspected: $imageTag"
    }
}

$targetImage = "${imageTag}@$($imageInfo.Digest)"
if ($BootstrapPullSecret) {
    Write-Warning 'Bootstrapping the cluster pull secret from the current local CCR login. Prefer a dedicated read-only pull credential when the registry supports one.'
    $credential = Get-DockerCredential
    try {
        Sync-CcrPullSecret -Credential $credential
    }
    finally {
        $credential = $null
    }
}
Assert-CcrPullSecret

$previousState = Get-RemoteDeploymentState
Write-Host "Previous:   $($previousState.Image)"
Write-Host "Target:     $targetImage"

$patchApplied = $false
try {
    Set-RemoteImage -TargetImage $targetImage -ExpectedImage $previousState.Image -ExpectedResourceVersion $previousState.ResourceVersion
    $patchApplied = $true
    Wait-RemoteRollout -TargetImage $targetImage
    Test-RemoteDeployment -TargetImage $targetImage
    if (-not (Test-PublicHealth)) {
        throw "Public health check failed: $PublicHealthUrl"
    }
}
catch {
    $deploymentError = $_
    if ($patchApplied) {
        try {
            Restore-PreviousImage -FailedImage $targetImage -PreviousImage $previousState.Image
        }
        catch {
            Write-Warning "Automatic rollback also failed: $($_.Exception.Message)"
        }
    }
    else {
        Write-Warning 'The atomic Deployment patch did not report success; automatic rollback was skipped to avoid reverting a concurrent deployment.'
    }
    throw $deploymentError
}

Write-Host ''
Write-Host 'Deployment succeeded.'
Write-Host "Image: $targetImage"
Write-Host "Rollback example: .\scripts\deploy-backend.ps1 -RollbackTag $targetTag"
