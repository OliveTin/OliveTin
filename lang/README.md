Hey, thanks for reading this quick introduction to translations. The project founder
only speaks two languages; English, and Bad English (!), so the initial translations
have all been AI-generated. It is assumed that "something is better than nothing". 

It would be most welcome to have human contributors who are native speakers improve these
translations, or add new ones.

## How to contribute

If a translation file does not exist for your locale, you can create one by copying
en.yaml and changing the locale code in the filename. You can view the language that
your browser reports by opening the "Select Language" dialog from the footer.

## File format

Internally, OliveTin uses the vue-i18n library for translations. This does support
language pluralization and other advanced features. For docs, check the following;

https://vue-i18n.intlify.dev/guide/essentials/pluralization.html

The translation files are in YAML format. Each file contains key-value pairs. 

OliveTin developers then "process" these files into JSON format used for the app.

If you are able, it would be appreciated if you run `make` in the language directory 
to process your language file before submitting a PR. This will ensure that the JSON
file is up to date. If you don't understand how to do this, don't worry; just submit
the YAML file and the developers will take care of it.

## Contributing improvements

Please check out the file `CONTRIBUTING.md` for instructions on how to submit a pull 
request with your improvements.

As always, if you need any help, please feel free to raise an issue on GitHub or 
jump into the Discord server for OliveTin.
