#!/usr/bin/env python3

import glob
import re

adoc_files = glob.glob('pages/**/*.adoc', recursive=True)

filelist = list()

for file in adoc_files:
    with open(file, 'r') as handle:
        content = handle.read()

        matches = re.findall('^= ', content, re.MULTILINE)

        if len(matches) == 0:
            filelist.append(file)


print("Files:", len(filelist))

for file in filelist:
    print(file)

