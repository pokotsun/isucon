ls
bash env.sh 
ls
cd isubata/
ls
cd ..
ls
sudo vim /etc/systemd/system/isubata.golang.service 
vim env.sh 
sudo vim /etc/systemd/system/isubata.golang.service 
exit
ls
cd isubata/
ls
vim README.md 
cd bench/
./bin/bench -remotes=127.0.0.1 -output result.json
jq . < result.json 
sudo apt install jq
jq . < result.json 
exit
