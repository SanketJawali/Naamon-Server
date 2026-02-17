from fastapi import FastAPI
from pathlib import Path

app = FastAPI()


@app.get("/")
def home():
    return {"msg": "Hello World"}


@app.get("/blog/{slug}")
def read_item(slug: str):
    blog_path = Path("blogs") / f"{slug}.md"
    content = blog_path.read_text()
    return {"slug": slug, "content": content}
