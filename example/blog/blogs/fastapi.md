# FastAPI: Modern API Development with Python

FastAPI is a modern web framework for building APIs with Python. It's designed to be fast, easy to use, and easy to learn. By leveraging Python type hints and async/await syntax, FastAPI automatically generates interactive API documentation and provides exceptional performance. If you're looking to build robust APIs quickly, FastAPI is an excellent choice.

## Why FastAPI?

FastAPI offers compelling advantages:
- **Automatic API documentation**: OpenAPI and Swagger UI generated automatically
- **Performance**: One of the fastest Python frameworks, comparable to Node.js and Go
- **Type hints**: Built-in validation through Python's type system
- **Async support**: Native async/await for handling concurrent requests
- **Easy to learn**: Clean, intuitive syntax following modern Python standards
- **Production-ready**: Used by companies like Netflix, Uber, and Microsoft

## Installation

Create a virtual environment and install FastAPI with Uvicorn (ASGI server):

```bash
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate

pip install fastapi uvicorn
```

## Your First API

Create `main.py`:

```python
from fastapi import FastAPI

app = FastAPI()

@app.get("/")
def read_root():
    return {"message": "Hello, FastAPI!"}

@app.get("/hello/{name}")
def greet(name: str):
    return {"greeting": f"Hello, {name}!"}
```

Run the server:
```bash
uvicorn main:app --reload
```

Visit:
- Main API: `http://localhost:8000/`
- Swagger UI docs: `http://localhost:8000/docs`
- ReDoc docs: `http://localhost:8000/redoc`

## Path Parameters and Query Parameters

FastAPI makes it easy to extract parameters:

```python
from fastapi import FastAPI

app = FastAPI()

# Path parameter
@app.get("/users/{user_id}")
def get_user(user_id: int):
    return {"user_id": user_id}

# Query parameters
@app.get("/search")
def search(query: str, limit: int = 10, skip: int = 0):
    return {
        "query": query,
        "limit": limit,
        "skip": skip
    }

# Both path and query
@app.get("/posts/{post_id}/comments")
def get_comments(post_id: int, sort_by: str = "date"):
    return {
        "post_id": post_id,
        "sort_by": sort_by
    }
```

## Request and Response Models

Use Pydantic models for validation and documentation:

```python
from fastapi import FastAPI
from pydantic import BaseModel

app = FastAPI()

class Item(BaseModel):
    name: str
    price: float
    is_available: bool = True

class User(BaseModel):
    name: str
    email: str
    age: int

@app.post("/items/")
def create_item(item: Item):
    return {
        "created": item,
        "total_cost": item.price * 1.1  # 10% markup
    }

@app.get("/users/{user_id}", response_model=User)
def get_user(user_id: int):
    # Fetch user from database
    return {
        "name": "Alice",
        "email": "alice@example.com",
        "age": 28
    }
```

## HTTP Methods

FastAPI supports all HTTP methods:

```python
# GET - Retrieve data
@app.get("/items")
def list_items():
    return [{"id": 1, "name": "Item 1"}]

# POST - Create data
@app.post("/items")
def create_item(item: Item):
    return {"created": item, "id": 1}

# PUT - Replace data
@app.put("/items/{item_id}")
def update_item(item_id: int, item: Item):
    return {"id": item_id, "item": item}

# DELETE - Remove data
@app.delete("/items/{item_id}")
def delete_item(item_id: int):
    return {"deleted": item_id}

# PATCH - Partial update
@app.patch("/items/{item_id}")
def partial_update(item_id: int, name: str = None):
    return {"id": item_id, "updated_name": name}
```

## Status Codes and Headers

Control response behavior:

```python
from fastapi import FastAPI, status
from typing import Optional

app = FastAPI()

@app.post("/items", status_code=status.HTTP_201_CREATED)
def create_item(item: Item):
    return item

@app.get("/items", headers={"X-Total-Count": "100"})
def list_items():
    return [{"id": 1}]
```

## Error Handling

```python
from fastapi import FastAPI, HTTPException

app = FastAPI()

users = {1: "Alice", 2: "Bob"}

@app.get("/users/{user_id}")
def get_user(user_id: int):
    if user_id not in users:
        raise HTTPException(
            status_code=404,
            detail="User not found"
        )
    return {"name": users[user_id]}
```

## Async Support

FastAPI's async support enables handling many concurrent requests:

```python
import asyncio
from fastapi import FastAPI

app = FastAPI()

@app.get("/items")
async def list_items():
    # Simulate database query
    await asyncio.sleep(1)
    return [{"id": 1, "name": "Item 1"}]

@app.get("/data")
async def fetch_data():
    # Can call multiple async operations concurrently
    results = await asyncio.gather(
        get_users(),
        get_posts()
    )
    return results

async def get_users():
    await asyncio.sleep(0.5)
    return [{"id": 1, "name": "Alice"}]

async def get_posts():
    await asyncio.sleep(0.5)
    return [{"id": 1, "title": "First Post"}]
```

## Dependency Injection

FastAPI's dependency system promotes code reuse:

```python
from fastapi import FastAPI, Depends
from typing import Optional

app = FastAPI()

def get_token(token: Optional[str] = None):
    if token is None:
        raise HTTPException(status_code=401, detail="No token")
    return token

@app.get("/protected")
def protected_route(token: str = Depends(get_token)):
    return {"message": "You have access!", "token": token}

# Query parameter dependency
def pagination(skip: int = 0, limit: int = 10):
    return {"skip": skip, "limit": limit}

@app.get("/items")
def list_items(pag: dict = Depends(pagination)):
    return {"items": [], "pagination": pag}
```

## Security Basics

```python
from fastapi import FastAPI, HTTPException, status
from fastapi.security import HTTPBasic, HTTPBasicCredentials

app = FastAPI()
security = HTTPBasic()

@app.get("/secure")
def secure_route(credentials: HTTPBasicCredentials = Depends(security)):
    if credentials.username == "admin" and credentials.password == "secret":
        return {"message": "Welcome, admin!"}
    raise HTTPException(
        status_code=status.HTTP_401_UNAUTHORIZED,
        detail="Incorrect credentials"
    )
```

## A Complete Example

```python
from fastapi import FastAPI, HTTPException, status
from pydantic import BaseModel

app = FastAPI(title="Blog API")

class Post(BaseModel):
    id: int
    title: str
    content: str
    author: str

posts_db = [
    Post(id=1, title="FastAPI Intro", content="FastAPI is awesome", author="Alice"),
    Post(id=2, title="Python Tips", content="Tips and tricks", author="Bob")
]

@app.get("/posts", response_model=list[Post])
def get_posts():
    return posts_db

@app.get("/posts/{post_id}", response_model=Post)
def get_post(post_id: int):
    for post in posts_db:
        if post.id == post_id:
            return post
    raise HTTPException(status_code=404, detail="Post not found")

@app.post("/posts", response_model=Post, status_code=status.HTTP_201_CREATED)
def create_post(post: Post):
    posts_db.append(post)
    return post

@app.delete("/posts/{post_id}")
def delete_post(post_id: int):
    for i, post in enumerate(posts_db):
        if post.id == post_id:
            posts_db.pop(i)
            return {"message": "Post deleted"}
    raise HTTPException(status_code=404, detail="Post not found")
```

## Deployment

Deploy FastAPI applications to Heroku, Railway, AWS, or any cloud platform:

```bash
# Create requirements.txt
pip freeze > requirements.txt

# Build Docker image
docker build -t my-api .

# Run container
docker run -p 80:8000 my-api
```

## Best Practices

1. **Use type hints everywhere**: Improves documentation and catches errors
2. **Organize with routers**: Split large APIs into modules
3. **Validate input**: Leverage Pydantic models
4. **Handle errors gracefully**: Provide meaningful error messages
5. **Use async for I/O**: Database calls, API requests
6. **Test thoroughly**: FastAPI works well with pytest

## Next Steps

- Add database integration with SQLAlchemy
- Implement JWT authentication
- Build comprehensive APIs with multiple endpoints
- Deploy to production
- Explore WebSockets for real-time communication

## Conclusion

FastAPI makes building modern, high-performance APIs in Python incredibly enjoyable. Its combination of ease of use, automatic documentation, and excellent performance makes it perfect for both beginners and experienced developers. Start building your next API with FastAPI!
