import json
import os
import re
import datetime
import pymongo
from bson.objectid import ObjectId

import helper
import handler_basic

mgo_cli_ev = None
mgo_db_ev = None

COLLNAME_SRC = "news"
COLLNAME_CRAWLING = "crawling"
COLLNAME_CRAWLING_ENTITY = "crawling_entity"

def process(newsId):
    coll_src = helper.mongo.get_collection_cr(COLLNAME_SRC)
    coll_crawling = helper.mongo.get_collection_nt(COLLNAME_CRAWLING)
    coll_crawling_entity = helper.mongo.get_collection_nt(COLLNAME_CRAWLING_ENTITY)
    
    news = coll_src.find_one({"newsId":newsId})
    if news is None:
        helper.eventlog.error("Crawling News Not Found: %s")
        return
    
    category = re.findall('\"category\":\"\s*((?:.|\n)*?)\"', news["mobileNews"])
    # print(''.join(category))
    coll_crawling.update_one({"newsId":newsId}, {"$set":{"category":''.join(category)}}, upsert=True)
    
def test():
    coll = helper.mongo.get_collection_cr(COLLNAME_SRC)
    docs = coll.find({}, {"newsId":1}).limit(5)    
    # docs = coll.find({"newsId":"20180823202045147"}, {"newsId":1})    
    for doc in docs:
        process(doc["newsId"])
    print("DONE+")

def getRaw(): #크롤링원문(html) 복사
    coll = helper.mongo.get_collection_cr(COLLNAME_SRC)
    coll_crawling = helper.mongo.get_collection_nt(COLLNAME_CRAWLING)

    # docs = coll.find({}, {"newsId":1 , "title":1 , "mobileNews":1})#.limit(5) 
    docs = coll.find({}, {"newsId":1 , "title":1 , "cpKorName":1})#.limit(5) 
    for doc in docs:   
        coll_crawling.update_one({"newsId":doc['newsId']}, {"$set":{"mediaName":doc['cpKorName']}}, upsert=True)
        # print(doc['newsId']+' \ '+doc['title'])
    print('DONE')

def getSampleCSV():
    # cats = ['사회','정치','경제','국제','문화','연예일반','스포츠일반']
    # cats = ['축구', '야구']
    cats = ['IT 과학', '교육']
    coll_crawling = helper.mongo.get_collection_nt(COLLNAME_CRAWLING)
    t = ""
    f = open('tests/csvfile3.csv','w',encoding='utf8')
    for cat in cats:
        docs = coll_crawling.find({
            "categoryCalc":cat
        })
        # rgx = re.compile('"category":"' + cat, re.IGNORECASE)
        # if cat == "국제":
        #     docs = coll_crawling.find({
        #         "$expr": { "$gt": [ { "$strLenCP": "$content" }, 600 ] }, 
        #         "raw":rgx,
        #         "title":{"$regex":"[가-힣]"}
        #     }).sort([("newsId", -1)]).limit(100)
        # else:
        #     docs = coll_crawling.find({
        #         "$expr": { "$gt": [ { "$strLenCP": "$content" }, 600 ] }, 
        #         "raw":rgx,
        #         # "title":{"$regex":"^((?![(AG)(아시안게임)]).)*$"}
        #     }).sort([("newsId", 1)]).limit(100)

        for doc in docs:
            # print(cat)
            t += cat + "\t" + doc['mediaName'] + "\t" + doc['title'] \
            + "\thttps://media.daum.net/v/" + doc['newsId']  + "\t" + str(len(doc['content'])) +  "\n"

    print(t)
    f.write(t)
    f.close()

def setCategoryML():    
    coll_crawling = helper.mongo.get_collection_nt(COLLNAME_CRAWLING)

    for (dirpath, dirnames, filenames) in os.walk("tests/data"):
        for filename in filenames:
            if not filename.endswith(".csv"):
                continue
            print(filename)
            csv = os.path.join(dirpath, filename)
            inf = open(csv, encoding="utf-8")
            for line in inf.readlines():
                t = line.rstrip().split("|")
                print(t[1] + "..." +  t[3])
                coll_crawling.update_one({"newsId":t[1]}, {"$set":{"categoryCalc":t[3]}}, upsert=True)
    print("DONE.")

def connect_ev():
    global mgo_cli_ev, mgo_db_ev
    mgo_cli_ev = pymongo.MongoClient("mongodb://172.17.0.1:3013")
    mgo_db_ev = mgo_cli_ev["newstrust_201810"]
    mgo_cli_ev.server_info() # will raise exception if failed
    print("Successfully connected to MongoDB")

def setEvaluation():
    connect_ev()
    coll_a = mgo_db_ev["assessments"]
    coll_b = mgo_db_ev["assessment_news"]
    coll_crawling = helper.mongo.get_collection_nt(COLLNAME_CRAWLING)

    docs = coll_a.aggregate(
        [
            {
                "$group" : {
                    "_id" : "$assessment_news_id",
                    "count": { "$sum": 1 },
                    "balance" : { "$avg": "$balance" }, 
                    "variety" : { "$avg": "$variety" }, 
                    "uniqueness" : { "$avg": "$uniqueness" }, 
                    "reality" : { "$avg": "$reality" }, 
                    "importance" : { "$avg": "$importance" }, 
                    "clariry" : { "$avg": "$clariry" }, 
                    "deep" : { "$avg": "$deep" }, 
                    "readability" : { "$avg": "$readability" }, 
                    "inflammation" : { "$avg": "$inflammation" }, 
                    "usefulness" : { "$avg": "$usefulness" }, 
                    "average" : { "$avg": "$average" }
                }
            }
        ]
    )
    for doc in docs:
        newsId = coll_b.find_one({"_id":ObjectId(doc["_id"])})["newsId"]
        print("%s : %s / balance : %s" % (doc["_id"], newsId, doc["balance"]))
        coll_crawling.update_one(
            {"newsId" : newsId},
            {"$set" : {
                "evaluation" : {
                    "balance" : doc["balance"], 
                    "variety" : doc["variety"], 
                    "uniqueness" : doc["uniqueness"], 
                    "reality" : doc["reality"], 
                    "importance" : doc["importance"], 
                    "clariry" : doc["clariry"], 
                    "deep" : doc["deep"], 
                    "readability" : doc["readability"], 
                    "inflammation" : doc["inflammation"], 
                    "usefulness" : doc["usefulness"], 
                    "average" : doc["average"]
                }
            }},
            upsert=False
        )
    print("DONE")

if __name__ == "__main__":
    print("Do not execute this script standalone.")
    helper.mongo.connect_cr()
    helper.mongo.connect_nt()
    helper.eventlog.set_worker_id("worker_manual")
    # getRaw()
    # getSampleCSV()
    # setCategoryML()
    setEvaluation()
    