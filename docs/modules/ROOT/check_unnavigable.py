#!/usr/bin/env python3

# find .adoc files that are not navigable from the nav.adoc file

import glob

nav_file = open('nav.adoc', 'r')
nav_string = nav_file.read()

adoc_files = glob.glob('pages/**/*.adoc', recursive=True)

unnavigable_files = []

for file in adoc_files:
    filename = file.replace("pages/", "")

    if filename not in nav_string:
        unnavigable_files.append(filename)


unnavigable_files = sorted(unnavigable_files)

print("Unnavigable files:", len(unnavigable_files))
for file in unnavigable_files:
    print(file)
