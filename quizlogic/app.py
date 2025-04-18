import json
import logging
import sys
import psutil
from quiz_generation import generate_quiz

# Setup logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(levelname)s - %(message)s",
)
logger = logging.getLogger(__name__)

def main():
    try:
        # Log memory usage before processing
        process = psutil.Process()
        mem_info = process.memory_info()
        logger.info(f"Memory usage before quiz generation: {mem_info.rss / 1024 / 1024:.2f} MB")

        # Read JSON input from stdin
        input_data = sys.stdin.read().strip()
        if not input_data:
            logger.error("No input data received")
            print(json.dumps({"ok": False, "data": ["No input data provided"]}), file=sys.stdout)
            sys.stdout.flush()
            return

        try:
            data = json.loads(input_data)
        except json.JSONDecodeError as e:
            logger.error(f"Invalid JSON input: {str(e)}")
            print(json.dumps({"ok": False, "data": [f"Invalid JSON: {str(e)}"]}), file=sys.stdout)
            sys.stdout.flush()
            return

        logger.info(f"Received quiz request: {data}")

        # Validation
        user_id = data.get("user_id")
        topic = data.get("topic")
        num_questions = data.get("num_questions")
        difficulty = data.get("difficulty", "").lower()

        if not (5 <= num_questions <= 20):
            logger.error(f"Invalid num_questions: {num_questions}")
            print(json.dumps({"ok": False, "data": ["Number of questions must be between 5 and 20"]}), file=sys.stdout)
            sys.stdout.flush()
            return

        if difficulty not in ["easy", "medium", "hard"]:
            logger.error(f"Invalid difficulty: {difficulty}")
            print(json.dumps({"ok": False, "data": ["Difficulty must be easy, medium, or hard"]}), file=sys.stdout)
            sys.stdout.flush()
            return

        # Generate quiz
        result = generate_quiz(data)
        logger.info(f"Generated quiz: {len(result.get('data', []))} questions")

        # Log memory usage after processing
        mem_info = process.memory_info()
        logger.info(f"Memory usage after quiz generation: {mem_info.rss / 1024 / 1024:.2f} MB")

        # Write JSON output to stdout
        print(json.dumps(result), file=sys.stdout)
        sys.stdout.flush()

    except Exception as e:
        logger.error(f"Unexpected error: {str(e)}", exc_info=True)
        print(json.dumps({"ok": False, "data": [f"Internal server error: {str(e)}"]}), file=sys.stdout)
        sys.stdout.flush()

if __name__ == "__main__":
    main()