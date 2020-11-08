#/bin/sh

set -e

mkdir -p lib

# Build tekstitv library
cd third_party/yle-tekstitv
./configure --disable-completion\
            --disable-executable\
            --disable-shared-lib
make
cp build/lib/libtekstitv.a include/tekstitv.h ../../lib
