
# vib-fsguard

[Vib](https://github.com/vanilla-os/vib) plugin that sets up fsguard and generates a filelist.

This plugin requires that `minisign` is installed in the image, this can be easily done with a nested module

## Module Structure

```yaml
- name: fsguard
  type: fsguard
  CustomFsGuard: false
  FsGuardLocation: "/usr/bin/"
  GenerateKey: true
  FilelistPaths: ["/usr/bin/"]
  modules:
    - name: minisign
      type: apt
      sources:
        packages:
            - "minisign"
```

If `GenerateKey` is set to false, `KeyPath` has to be specified, pointing to a location in the container (e.g. added through includes.container) which contains already existing minisign keys:

```yaml
- name: fsguard
  type: fsguard
  CustomFsGuard: false
  FsGuardLocation: "/usr/bin/"
  GenerateKey: false
  KeyPath: "/etc/minisign/"
  FilelistPaths: ["/usr/bin/"]
  modules:
    - name: minisign
      type: apt
      sources:
        packages:
            - "minisign" 
```

Note that the keys must be named `minisign.pub` (public) and `minisign.key` (private) in this example the minisign keys would be in `includes.container/etc/minisign/`, which translates to `/etc/minisign** in the build environment

Keep in mind that the minisign key **cannot** be password protected, as there is no way for the user to type in the password during building (which is why always generting a random key through GenerateKey is recommended)

In the case that FsGuard has to be manually built (due to a configuration change or similiar), the `CustomFsGuard` option has to be set to True, this stops the module from fetching a prebuilt FsGuard and instead allows the user to manually build FsGuard, it does however expect the FsGuard binary to be at `/sources/FsGuard`

## Use of Generative AI

Maintainers may use generative AI tools as assistants while working on vib-fsguard. Non-trivial assisted commits disclose the tool, model, and scope of the work.

AI tools may assist with code comments, documentation, repetitive code, and issue triage. Maintainers make project decisions and review every assisted change before it is merged.

Use these trailers for non-trivial assisted commits:

```plain
Assisted-by: <tool>:<model-version>
AI-Scope: <what the tool generated and the prompt or a short prompt summary>
```

Single-line completions, renames, and formatting changes do not need trailers.

Coding agents must also follow [AGENTS.md](AGENTS.md) before changing files,
creating commits, or opening pull requests.
