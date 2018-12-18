import os
import logging
import pymongo

mgo_cli_nt = None
mgo_db_nt = None
mgo_cli_cr = None
mgo_db_cr = None

def connect_nt():
    global mgo_cli_nt, mgo_db_nt
    mgo_cli_nt = pymongo.MongoClient(os.environ["MONGO_URL_NTRUST"])
    mgo_db_nt = mgo_cli_nt[os.environ["MONGO_DB_NTRUST"]]
    mgo_cli_nt.server_info() # will raise exception if failed
    logging.info("Successfully connected to MongoDB %s", os.environ["MONGO_URL_NTRUST"])

def close_nt():
    global mgo_cli_nt, mgo_db_nt
    if mgo_cli_nt is not None:
       mgo_cli_nt.close()
       mgo_cli_nt = None
       mgo_db_nt = None

def get_collection_nt(collname):
    if mgo_db_nt is None:
        raise RuntimeError("MongoDB not connected")
    return mgo_db_nt[collname]

def connect_cr():
    global mgo_cli_cr, mgo_db_cr
    mgo_cli_cr = pymongo.MongoClient(os.environ["MONGO_URL_CRAWLING"])
    mgo_db_cr = mgo_cli_cr[os.environ["MONGO_DB_CRAWLING"]]
    mgo_cli_cr.server_info() # will raise exception if failed
    logging.info("Successfully connected to MongoDB %s", os.environ["MONGO_URL_CRAWLING"])

def close_cr():
    global mgo_cli_cr, mgo_db_cr
    if mgo_cli_cr is not None:
       mgo_cli_cr.close()
       mgo_cli_cr = None
       mgo_db_cr = None

def get_collection_cr(collname):
    if mgo_db_cr is None:
        raise RuntimeError("MongoDB not connected")
    return mgo_db_cr[collname]
