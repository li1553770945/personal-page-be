cd /srv/personal-page-be
if [ ! -d "./pidfile.txt" ]; then
  kill -9 `cat pidfile.txt`
fi

chmod +x ./personal-page-be
nohup ./personal-page-be > logfile.txt & echo $! > pidfile.txt