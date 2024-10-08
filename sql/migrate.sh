psql -h localhost -U leaderboard leaderboard -c "COPY mdw TO STDOUT WITH CSV HEADER" | psql -h ep-dawn-frog-a4kb9tw2-pooler.us-east-1.aws.neon.tech  -U default verceldb -c "COPY mdw FROM STDIN WITH CSV HEADER"

