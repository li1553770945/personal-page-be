cd /srv/personal-page-be
if [ ! -d "./pidfile.txt" ]; then
  kill -9 `cat pidfile.txt`
fi

chmod +x ./main
nohup ./main > logfile.txt & echo $! > pidfile.txt