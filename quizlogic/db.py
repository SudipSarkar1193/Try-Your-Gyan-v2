import psycopg2
import os
import dotenv
from dotenv import load_dotenv
import logging

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Load .env
load_dotenv()

# Database configuration
DB_CONFIG = {
    'user': os.getenv("user"),
    'password': os.getenv("password"),
    'host': os.getenv("host"),
    'port': os.getenv("port"),
    'dbname': os.getenv("dbname")
}

class DatabaseConnection:
    def __init__(self):
        self.connection = None
    
    def __enter__(self):
        try:
            self.connection = psycopg2.connect(**DB_CONFIG)
            logger.info("Database connection established")
            return self.connection
        except Exception as e:
            logger.error(f"Failed to connect to database: {e}")
            raise

    def __exit__(self, exc_type, exc_val, exc_tb):
        if self.connection:
            self.connection.close()
            logger.info("Database connection closed")

def get_past_questions(user_id: int, topic: str) -> list:
    """Fetch past questions for a user and topic (synchronous version)"""
    logger.info(f"Fetching past questions for user {user_id} and topic {topic}")
    
    try:
        with DatabaseConnection() as conn:
            with conn.cursor() as cur:
                cur.execute(
                    """
                    SELECT q.question 
                    FROM questions q
                    JOIN quizzes z ON q.quiz_id = z.id
                    WHERE z.user_id = %s AND LOWER(z.quiz_name) = LOWER(%s)
                    """,
                    (user_id, topic)
                )
                past_questions = [row[0] for row in cur.fetchall()]
                logger.info(f"Found {len(past_questions)} past questions")
                return past_questions
                
    except Exception as e:
        logger.error(f"Database error for user {user_id}, topic {topic}: {str(e)}")
        return []  # Return empty list instead of raising exception for better resilience

# Test function (optional)
def test_db_connection():
    try:
        with DatabaseConnection() as conn:
            with conn.cursor() as cur:
                cur.execute("SELECT version();")
                version = cur.fetchone()
                logger.info(f"Database connection test successful. PostgreSQL version: {version[0]}")
        return True
    except Exception as e:
        logger.error(f"Database connection test failed: {e}")
        return False