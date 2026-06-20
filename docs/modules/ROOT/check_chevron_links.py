#!/usr/bin/env python3

import glob
import re

nav_file = open('nav.adoc', 'r')
nav_string = nav_file.read()

adoc_files = glob.glob('pages/**/*.adoc', recursive=True)

filelist = dict()

for file in adoc_files:
    with open(file, 'r') as handle:
        content = handle.read()

        matches = re.findall(r'<<(.*?),?([\w\- ]+)>>', content)

        for match in matches:
            m = match

            if match[0] == "":
                m = match[1]
            else:
                m = match[0]

            if content.count("#" + m) != 1:
                if file not in filelist:
                    filelist[file] = list()

                filelist[file].append(m)


print("Files:", len(filelist))

for file in filelist.keys():
    print(file)

    for match in filelist[file]:
        print("\t", match)
