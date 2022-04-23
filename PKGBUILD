pkgname=bserv
pkgver=0
pkgrel=1
pkgdesc='Simple backup server which stores uploaded files.'
arch=('x86_64')
url='https://git.vimusov.space/me/bserv'
license=('GPL')
makedepends=('go' 'make')
source=("${pkgname}.go" makefile)
md5sums=('SKIP' 'SKIP')

pkgver()
{
    printf "r%s.%s" "$(git rev-list --count HEAD)" "$(git rev-parse --short HEAD)"
}

build()
{
    make -C "$srcdir"
}

package()
{
    make -C "$srcdir" DESTDIR="$pkgdir" install
}
