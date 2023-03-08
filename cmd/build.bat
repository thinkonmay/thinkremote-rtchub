rmdir /q /s build
rmdir /q /s lib
mkdir build && cd build 



cmake .. -G "Ninja"
ninja

cd ..
robocopy build listener/cgo/lib liblistener.a
robocopy build broadcaster/cgo/lib libbroadcaster.a
