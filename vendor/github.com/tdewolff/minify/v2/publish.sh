#!/bin/sh

VERSION=`git describe --tags --abbrev=0 | cut -c 2-`

cd dist
echo ""
echo "Downloading v$VERSION.tar.gz..."
wget -q --show-progress https://github.com/tdewolff/minify/archive/v$VERSION.tar.gz
SHA256=`sha256sum v$VERSION.tar.gz`
SHA256=( $SHA256 )

echo ""
echo "Releasing for AUR..."
cd /home/taco/dev/aur/minify
sed -i "s/^pkgver=.*$/pkgver=$VERSION/" PKGBUILD
sed -i "s/^sha256sums=.*$/sha256sums=('$SHA256')/" PKGBUILD
./build.sh
git commit -am "Update to v$VERSION"
git push
cd -

echo ""
echo "Releasing for Homebrew..."
cd /home/taco/dev/brew/homebrew-tap/Formula
sed -i "s,^  url \".*\"$,  url \"https://github.com/tdewolff/minify/archive/v$VERSION.tar.gz\"," minify.rb
sed -i "s/^  sha256 \".*\"$/  sha256 \"$SHA256\"/" minify.rb
git commit -am "Update to v$VERSION"
git push
cd -

#echo ""
#echo "Releasing Python bindings..."
#cd ../bindings/py
#make publish
#cd -
