echo "don't forget to do export GITHUB_TOKEN=...";
export BOOT2KEVOREECLIVERSION="v0.9.9";
github-release release -u kevoree -r boot2kevoree-cli --tag $BOOT2KEVOREECLIVERSION --name "Boot2Kevoree" --description "Command line utility to manipulate Boot2Kevoree" --pre-release

github-release release -u kevoree -r boot2kevoree-cli --tag $BOOT2KEVOREECLIVERSION --name "boot2kevoree_mac" --file build/boot2kevoree_mac.zip
github-release release -u kevoree -r boot2kevoree-cli --tag $BOOT2KEVOREECLIVERSION --name "boot2kevoree_linux" --file boot2kevoree-cli_linux_amd64
github-release release -u kevoree -r boot2kevoree-cli --tag $BOOT2KEVOREECLIVERSION --name "boot2kevoree.exe" --file boot2kevoree-cli_windows_amd64.exe


github-release delete -u kevoree -r boot2kevoree-cli --tag $BOOT2KEVOREECLIVERSION