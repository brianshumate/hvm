## Informal hvm Roadmap

This is an informal project roadmap list of items needing attention and help:

### Pre-Release Roadmap

These items need attention before the first release is possible:

1. Complete uninstall command
2. Complete list command
  - `list` command shows locally installed versions by default
  - `list -remote` will query releases.hashicorp.com and list versions found there
3. Fix issues with current version detection in use command
4. Build/update release system scripting that produces results compatible with GH releases

### Post-Release Roadmap

1. Add convenience options like `install --all` to get all latest binary versions in one shot
