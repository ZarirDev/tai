# Maintainer: Ahmad Zarir <self@zarir.org>
pkgname=tai
pkgver=1.0.0
pkgrel=1
pkgdesc="CLI chatbot for Groq with web search"
arch=('x86_64' 'aarch64')
url="https://github.com/zarirdev/tai"
license=('MIT')
depends=('glibc')
makedepends=('go')
source=("$pkgname-$pkgver.tar.gz::$url/archive/refs/tags/v$pkgver.0.tar.gz")
sha256sums=('d5558cd419c8d46bdc958064cb97f963d1ea793866414c025906ec15033512ed') # replace with actual hash after release

build() {
  cd "$pkgname-$pkgver"
  go build -ldflags "-s -w" -o tai .
}

package() {
  cd "$pkgname-$pkgver"

  # Install binary
  install -Dm755 tai "$pkgdir/usr/bin/tai"

  # Install system-wide default config (read-only fallback)
  install -Dm644 /dev/null "$pkgdir/etc/tai/config.yaml"
  cat > "$pkgdir/etc/tai/config.yaml" <<EOF
# System-wide tai configuration (overriden by ~/.config/tai/config.yaml)
model: "groq/compound-mini"
max_tokens: 2048
debug: false
include_domains: []
exclude_domains: []
# api_key_encrypted is set per-user via 'tai --api'
EOF

  # Install docs (optional)
  install -Dm644 README.md "$pkgdir/usr/share/doc/tai/README.md"
  install -Dm644 LICENSE "$pkgdir/usr/share/licenses/tai/LICENSE"
}