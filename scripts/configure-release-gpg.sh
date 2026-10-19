#!/usr/bin/env bash
# Import the repository signing secret into the job's temporary keyring.
set -euo pipefail
: "${GPG_PRIVATE_KEY:?Configure the GPG_PRIVATE_KEY repository secret}"
: "${GNUPGHOME:?Set a temporary GNUPGHOME for this job}"
: "${RUNNER_TEMP:?RUNNER_TEMP is required}"

mkdir -m 700 -p "$GNUPGHOME"
printf '%s' "$GPG_PRIVATE_KEY" | gpg --batch --import
key=$(gpg --list-secret-keys --with-colons | awk -F: '$1 == "fpr" && !found {print $10; found=1}')
test -n "$key"
owner=$(gpg --list-keys --with-colons "$key" | awk -F: '$1 == "uid" && !found {print $10; found=1}')
email="${owner##*<}"; email="${email%>*}"
[[ "$email" == *@* ]] || { echo "::error::GPG key must contain your verified GitHub email"; exit 1; }
git config user.name "${owner% <*}"
git config user.email "$email"
git config user.signingkey "$key"
git config gpg.format openpgp
git config commit.gpgsign true
cat > "$RUNNER_TEMP/release-gpg" <<'SH'
#!/usr/bin/env bash
exec gpg --batch --pinentry-mode loopback --passphrase-fd 3 "$@" 3<<<"${GPG_PASSPHRASE:-}"
SH
chmod 700 "$RUNNER_TEMP/release-gpg"
git config gpg.program "$RUNNER_TEMP/release-gpg"
