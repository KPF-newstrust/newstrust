#!/bin/bash

cd mecab
wget newstrust-web:8091/public/userdic.csv

cd /opt/mecab/mecab-ko-dic*
USERDIC=/work/ntrust-worker/mecab/userdic.csv
echo "USERDIC * Script.."
if [ ! -f $USERDIC ]
then
  echo "USERDIC * Not Found"
elif [ -f $USERDIC ] && [ -f user-dic/newstrust.csv ] && [ -z "$(diff $USERDIC user-dic/newstrust.csv -q)" ]
then
  echo "USERDIC * No Changes."
else
  cp $USERDIC user-dic
  mv user-dic/userdic.csv user-dic/newstrust.csv
  cd tools
  ./add-userdic.sh
  echo "USERDIC * DONE - ./add-userdic.sh"
  cd ..
  make install
  echo "USERDIC * Done - make install."
fi
rm $USERDIC
