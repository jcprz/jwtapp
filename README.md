# jwtapp

This is from a JWT course app I took on Udemy. Just tiny tweaks to make it feel just a bit more "prod-like" for testing it.

The app will require these environment variables:

```
APP_PORT
DB_HOST
DB_USER
DB_PASSWORD
DB_PORT
DB_NAME
DB_DIALECT
REDIS_HOST
REDIS_PORT
REDIS_PASSWORD
SECRET
```

Most of them are self-explanatory so I'll skip to the less self-explanatory ones:
APP_PORT = the port that will listen on
DB_DIALECT = the dialect the app will talk (either "postgres" or "mysql", beware though that I only fully tested postgres
SECRET = this is needed for the token verification
