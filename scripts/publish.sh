#!/usr/bin/env bash
# Publish vibe-devkit to npm registry
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

echo "==> vibe-devkit publish helper"
echo ""

if ! command -v npm &>/dev/null; then
  echo "ERROR: npm not found. Install Node.js >= 18."
  exit 1
fi

if ! npm whoami &>/dev/null; then
  echo "Not logged in to npm."
  echo ""
  echo "Run: npm login"
  echo "Then re-run: ./scripts/publish.sh"
  exit 1
fi

echo "Logged in as: $(npm whoami)"
echo "Package: $(node -p "require('./package.json').name")@$(node -p "require('./package.json').version")"
echo ""

npm test
echo ""
npm pack --dry-run
echo ""

cat <<'EOF'
⚠️  npm yêu cầu 2FA để publish (lỗi E403 nếu thiếu).

Chọn 1 trong 2 cách:

A) Bật 2FA trên tài khoản npm (khuyến nghị)
   1. https://www.npmjs.com/settings/~/account → Two-Factor Authentication → Enable
   2. npm logout && npm login  (nhập OTP khi publish)

B) Dùng Granular Access Token (CI / không muốn OTP mỗi lần)
   1. https://www.npmjs.com/settings/~/tokens → Generate New Token → Granular
   2. Permissions: Read and Write, Packages: @hamanhtuan/vibe-devkit
   3. Bật "Bypass two-factor authentication for write actions"
   4. npm config set //registry.npmjs.org/:_authToken=YOUR_TOKEN
   5. Chạy lại: npm publish --access public

EOF

read -r -p "Publish to npm? [y/N] " confirm
if [[ "${confirm,,}" != "y" ]]; then
  echo "Aborted."
  exit 0
fi

if ! npm publish --access public; then
  echo ""
  echo "❌ Publish failed."
  echo ""
  echo "Nếu thấy E403 + 'Two-factor authentication or granular access token':"
  echo "  → Bật 2FA (cách A) hoặc dùng token có bypass 2FA (cách B) ở trên."
  exit 1
fi

echo ""
echo "✅ Published!"
echo ""
echo "Verify: npm view @hamanhtuan/vibe-devkit"
echo ""
echo "Users install with:"
echo "  npx @hamanhtuan/vibe-devkit install --preset flutter-supabase ."
