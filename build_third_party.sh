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
cd ../../lib
# Golang has type as a reserved word so we need to this
sed -i 's/html_item_type type/html_item_type i_type/g' tekstitv.h
sed -i 's/item\.type/item.\i_type/g' tekstitv.h
