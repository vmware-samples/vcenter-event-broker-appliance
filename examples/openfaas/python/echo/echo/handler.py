# Original function contribution by Michael Gasch https://github.com/embano1/of-echo/
def handle(event, context):
    print(event.body,flush=True)
    return {
        "statusCode": 200,
        "body": ""
    }
