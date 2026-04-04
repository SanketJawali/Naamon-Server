from fastapi import FastAPI
import time

app = FastAPI()


@app.get("/")
def index():
    return {
        "msg": "Hello World! This is a test API."
    }


@app.get("/delay/{duration}")
def delay(duration: int):
    try:
        if int(duration) > 0:
            time.sleep(duration)
            return {
                "msg": f"Delayed for {duration} seconds."
            }
        return {
            "msg": "Delayed for seconds."
        }

    except Exception as e:
        print(f"Error occured {e}")
        return {
            "err": "Error occured"
        }
