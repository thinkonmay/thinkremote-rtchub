rmdir /q /s build
rmdir /q /s lib
mkdir build && cd build 
cmake .. -G "Ninja"
ninja

cd ..
mkdir lib
robocopy build lib libshared.a
