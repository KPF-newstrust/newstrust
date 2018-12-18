import json
import os

import helper
import pika

COLLNAME_DIC = "dictionary"
COLLNAME_DIC_CUSTOM = "dictionary_custom"

def add_userdic():
    renew = false
    coll = helper.mongo.get_collection_nt(COLLNAME_DIC_CUSTOM)
    docs = coll.find({"applied":True})
    if docs is None:
        helper.eventlog.error("Words not found: %s")
        return

    # 세종,,,,NNP,지명,T,세종,*,*,*,*,*
    csv = ""
    for doc in docs:
        word = doc["word"]
        tail = "T" if((ord(word[-1:])-44032)%28) else "F" #종성유무
        csv += word + ",,,,NNP,*," + tail + "," + word + ",*,*,*,*,*\n"
    # print("============csv============\n%s" % (csv))

    c_old = "mecab/custom.csv"
    c_new = "mecab/custom-write.csv"
    f1 = open(c_new, 'w+', encoding="utf-8") # w+ :덮어쓰기
    f1.write(csv)
    f1.close()

    # print(os.stat(c_old).st_size)
    # print(os.stat(c_new).st_size)
    # print( open(c_new, 'r', encoding='utf-8').read() == open(c_old, 'r', encoding='utf-8').read())
    if os.path.exists(c_old) and \
        os.stat(c_old).st_size == os.stat(c_new).st_size and \
        open(c_old,'r',encoding='utf-8').read() == open(c_new,'r',encoding='utf-8').read():
        # 기존파일이 존재 && 사이즈가 동일 && 내용도 동일
        print("handler_dic - No changes.")
    else:
        print("handler_dic - Write")
        f2 = open("mecab/custom.csv", 'w+', encoding="utf-8")
        f2.write(csv)
        f2.close()
        renew = True

    return renew

def publish_userdic():
    renew = add_userdic()

    if renew == True :
        message = json.dumps({"cmd": "update", "ver": 1})
        # helper.rabbit.rb_notice_channel.basic_publish(exchange='notice', routing_key='', body=message)
        # conn = pika.BlockingConnection(pika.ConnectionParametrs(host=os.environ["MQ_URL"]))
        conn = pika.BlockingConnection(pika.connection.URLParameters(os.environ["MQ_URL"]))
        chan = conn.channel()
        chan.exchange_declare(exchange='notice', exchange_type='fanout')
        chan.basic_publish(exchange='notice',
                           routing_key='',
                           body=message)
        conn.close()

    helper.eventlog.debug("hander_dic - Added words : %d" % len)

if __name__ == "__main__":
    print("Do not execute this script standalone.")
    helper.mongo.connect_nt()
    publish_userdic()