cd build && mkdir build
cmake .. -G "Ninja"
ninja

cd ..
mkdir lib
robocopy build lib libshared.a

go build ./cmd/agent