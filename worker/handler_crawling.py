import json
import os
import re
import datetime
import pymongo

import helper
import handler_basic
from ntrust import jnscore

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

    section = re.findall('<section[^>]*>\s*((?:.|\n)*?)<\/section>', news["mobileNews"])
    
    content = re.findall('<p[^>]*>(.*?)<\/p>', ''.join(section)) # p 태그 내용 추출
    content = re.sub('<img\s(.*?)>', '', '\n'.join(content)) # img 태그 내용 제거
    content = re.sub('(<span[^>]*>|<strong>|<\/span>|<\/strong>)', '', content) # strong 태그 제거
    
    category = re.findall('\"category\":\"\s*((?:.|\n)*?)\"', news["mobileNews"])

    try:
        chgs_news, chgs_entity = handler_basic.get_metric_update_dicts({
            # "mediaId": None,
            "title": news['title'],
            "content": content
        })

        chgs_news["$set"]["content"] = content
        chgs_news["$set"]["category"] = ''.join(category)
        chgs_news["$set"]["mediaName"] = news['cpKorName']
        chgs_news["$set"]["raw"] = news['mobileNews']
        chgs_news["$unset"]["byline_writer"] = 1

        # UPDATE
        coll_crawling.update_one({"newsId":newsId}, chgs_news, upsert=True)
        
        chgs_entity["newsId"] = newsId
        coll_crawling_entity.replace_one({"newsId":newsId}, chgs_entity, upsert=True)
    except Exception as ex:
        print("_-_크롤링기사 분석 오류 (%s,%s)\n_-_%s\n_-_🤔%s" % (newsId, news['title'], content ,str(ex)))
        helper.eventlog.debug("_-_크롤링기사 분석 오류 (%s,%s)\n_-_%s\n_-_🤔%s" % (newsId, news['title'], content ,str(ex)))
        pass

def evaluate(newsId):
    coll = helper.mongo.get_collection_nt(COLLNAME_CRAWLING)
    doc = coll.find_one({"newsId":newsId})
    
    try:
        jsco = jnscore.evaluate(doc["categoryCalc"], "신문", doc)
        coll.update_one(
            {"newsId": doc["newsId"]}, 
            {"$set":{
                "journal":jsco.journal,
                "journal_totalSum": jsco.journalSum,
                "vanilla": jsco.vanilla,
                "vanilla_totalSum": jsco.vanillaSum,
                "score": jsco.scores
            }}, 
            upsert=False)
        print("  Done %s: %s (%s)" % (doc["newsId"], doc["title"], jsco.journalSum))
    except Exception as ex:
        return 1

    return 0


def run_process():
    coll = helper.mongo.get_collection_cr(COLLNAME_SRC)
    docs = coll.find({}, {"newsId":1}).limit(5)    

    # 파싱 오류난 경우
    # coll = helper.mongo.get_collection_nt(COLLNAME_CRAWLING)
    # docs = coll.find({"content":{"$exists":False}}, {"newsId":1})    

    for doc in docs:
        process(doc["newsId"])

def run_evaluate():
    coll = helper.mongo.get_collection_nt(COLLNAME_CRAWLING)    
    docs = coll.find({},{"newsId":1}).limit(5)
    
    # 파싱 오류난 경우
    # coll = helper.mongo.get_collection_nt(COLLNAME_CRAWLING)
    # docs = coll.find({"journal":{"$exists":False}}, {"newsId":1})    

    fail = 0
    for doc in docs:
        fail += evaluate(doc["newsId"])
    print("DONE. fail : %d" % (fail))

if __name__ == "__main__":
    print("Do not execute this script standalone.")
    helper.mongo.connect_cr()
    helper.mongo.connect_nt()
    helper.eventlog.set_worker_id("worker_manual")
    run_evaluate()
    # run_process()
    print("*Done*")
    