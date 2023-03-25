rmdir /q /s build
rmdir /q /s lib
mkdir build && cd build 



cmake .. -G "Ninja"
ninja

cd ..
robocopy build cgo/lib libshared.a
