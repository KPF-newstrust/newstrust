## 배포

모든 시스템은 [도커](https://www.docker.com/) 환경 위에서 구현됩니다. 도커를 설치한 후 아래 명령어로 배포합니다.  
마지막에 입력하는 숫자는 실행시킬 워커의 개수입니다. 지정하는 만큼 컨테이너가 생성됩니다. 
```
docker-compose up -d --scale worker=2
```

1. 도커 컨테이너들이 모두 올라오면 `{호스트}:60000`로 접속 가능합니다.
2. 초기 관리자계정 로그인은 `admin / 0000test`로 가능합니다.
3. web 컨테이너는 go 환경에서 실행되며, 소스코드가 수정되면 gin에 의해 자동으로 재빌드 및 배포 됩니다.

## 기타 명령어

```
# mongo 컨테이너만 다시 실행하기
docker-compose up -d --no-deps --build mongo
# web 컨테이너만 다시 실행(이미지 재생성)
docker-compose up -d --no-deps --build --force-recreate web
# worker 이미지 재빌드, 컨테이너 재시작
sudo docker-compose build worker
docker-compose restart worker
```