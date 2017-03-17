all:
	docker build -t abac/logaudit-server .

push:
	docker push abac/logaudit-server

up:
	docker service create \
		--constraint "node.role==worker" \
		--env "APP_BIND=:8020" \
		--env "APP_DB=root:myM0Joc5NBcE@tcp(db.jn.gl:3306)/logaudit?parseTime=true" \
		--env "APP_INTERVAL=43 */1 * * *" \
		--env "APP_AWS_ACCESS_KEY=AKIAIAXOWCEYIHWXSPGA" \
		--env "APP_AWS_SECRET_KEY=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" \
		--env "APP_AWS_REGION=us-west-2" \
		--name logaudit-server \
		--publish 8020:8020 \
		--mode global \
		--with-registry-auth \
		abac/logaudit-server

down:
	docker service rm logaudit-server

update:
	docker service update--image logaudit-server:latest --with-registry-auth logaudit-server