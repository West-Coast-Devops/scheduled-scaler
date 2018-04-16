#!/usr/bin/env python
import yaml

print_format="| {parameter:<40}| | {default:<50}|"
def walk_dict(d,keys=[],depth=0):
    for k,v in sorted(d.items(),key=lambda x: x[0]):
        keys.append(k)
        if isinstance(v,dict):
            walk_dict(v,keys,depth+1)
        else:
            print(print_format.format(parameter='`{0}`'.format(".".join(keys)),default='`{0}`'.format(v)))
        keys.pop()

s = open("./values.yaml")
d = yaml.load(s)

walk_dict(d)
