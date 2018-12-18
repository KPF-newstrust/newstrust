#!/usr/bin/env python3
#-*- coding: utf-8 -*-
import os
import datetime
import sys
# import json
# import configparser
# import logging
# import logging.config

import helper

helper.mongo.connect_nt()
coll = helper.mongo.get_collection_nt("dictionary")

def write(csv, version, source, date):
    inf = open(csv, encoding="utf-8")
    i = 0
    for line in inf.readlines():
        t = line.rstrip().split(",")
        w = {
            'source' : source,
            'word': t[0],
            'w1': t[1],
            'w2': t[2],
            'w3': t[3],
            'tag': t[4],
            'meaning': t[5],
            'tail': t[6],  # 종성유무 : T/F
            'sound': t[7],
            'type': t[8],  # Inflect - 활용 / Compound - 복합명사 / Preanalysis - 기분석
            'tag_first': t[9],
            'tag_last': t[10],
            'structure': t[11],
        }
        # print("t", t)

        ex = coll.find_one({'source' : source, 'word' : t[0], 'tag' : t[4]})
        #new
        if ex is None:
            w['flag'] = version
            w['version'] = version
            w['version_history'] = [{'version':version, 'result':'Inserted', 'date':date}]
            coll.save(w)
        #existed
        else:
            if ex['version'] == version:
                print('Same version.', version, ex, w)
            elif ex['meaning'] != w['meaning'] or ex['tail'] != w['tail'] or \
                    ex['sound'] != w['sound'] or ex['type'] != w['type'] or \
                    ex['tag_first'] != w['tag_first'] or ex['tag_last'] != w['tag_last'] or ex['structure'] != w['structure']:
                ex.update(w)
                ex['flag'] += '|'+version+'U'
                ex['version'] = version
                ex['version_history'] += [{'version':version, 'result':'Updated', 'date':date}]
                coll.save(ex)
            else:
                ex['flag'] += '|'+version+'X'
                ex['version'] = version
                ex['version_history'] += [{'version':version, 'result':'Unchanged', 'date':date}]
                coll.save(ex)
            i += 1

    print('exists :', i)

mecab = "mecab/mecab-ko-dic-2.1.1-20180720"
for (dirpath, dirnames, filenames) in os.walk(mecab):
    for filename in filenames:
        if not filename.endswith(".csv"):
            continue
        csv = os.path.join(dirpath, filename)
        print(datetime.datetime.now(), filename)
        write(csv, mecab[-8:], filename[:-4], '20180814')
        print(datetime.datetime.now(), filename, 'Done.')

helper.mongo.close_nt()
