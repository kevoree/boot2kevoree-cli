gox ..


rm boot2kevoree-cli
mv boot2kevoree-cli_darwin_amd64 boot2kevoree
zip -r -X boot2kevoree_mac.zip boot2kevoree
rm boot2kevoree-cli


rm boot2kevoree-cli
mv boot2kevoree-cli_linux_amd64 boot2kevoree
zip -r -X boot2kevoree_linux.zip boot2kevoree
rm boot2kevoree-cli


rm boot2kevoree-cli
mv boot2kevoree-cli_windows_amd64.exe boot2kevoree.exe
zip -r -X boot2kevoree_windows.zip boot2kevoree.exe
rm boot2kevoree.exe

rm boot2kevoree-cli*
rm boot2kevoree
