# SignPath artifact configurations (reference only)

These XML files are **paste templates** for the SignPath web UI. SignPath does not load them from this repository.

## Do not upload these `.xml` files as artifact samples

If you use **Upload an artifact sample** on `windows-zip.xml` / `windows-msi.xml`, SignPath treats them as XML documents to sign and generates an `<xml-file>` configuration. That feature is not available on SignPath Foundation / Open Source, and you get an error like:

> feature which is not currently available (XML element name: 'xml-file')

## Correct setup

1. In the SignPath project, **Add** an artifact configuration.
2. Choose **Custom** (edit XML), not “upload sample” of these files.
3. Paste the full contents of `windows-zip.xml` or `windows-msi.xml`.
4. Set the slug to `windows-zip` or `windows-msi` (must match CI).

Optional: use **Upload an artifact sample** with a real `OliveTin-windows-amd64.zip` or `.msi` from a build, then trim the generated config to match these references.
