from langchain_google_genai import ChatGoogleGenerativeAI
from langchain_core.prompts import ChatPromptTemplate
from langchain.output_parsers import StructuredOutputParser, ResponseSchema
import random
import os
from db import get_past_questions
import json
import re
import logging
import sys

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    stream=sys.stderr,
    format="%(filename)s - %(asctime)s - %(levelname)s - %(message)s"
)
logger = logging.getLogger(__name__)

# Define the expected JSON structure
response_schemas = [
    ResponseSchema(name="ok", description="Success indicator", type="boolean"),
    ResponseSchema(
        name="data",
        description="Array of questions or error message",
        type="array"
    )
]
output_parser = StructuredOutputParser.from_response_schemas(response_schemas)

def normalize_topic(raw_topic: str) -> str:
    """Normalize a raw topic string into a clean version"""
    topic = raw_topic.lower().strip()
    phrases_to_remove = [
        r"generate\s+me\s+a\s+quiz\s+on",
        r"create\s+a\s+quiz\s+about",
        r"quiz\s+on",
        r"a\s+quiz\s+on"
    ]
    for phrase in phrases_to_remove:
        topic = re.sub(phrase, "", topic)
    return topic.strip().title() or "Unknown"

def generate_quiz(request_data):
    try:
        logger.info(f"Starting generate_quiz with request: {request_data}")
        
        # Extract parameters from request
        user_id = request_data.get('user_id')
        topic = request_data.get('topic')
        num_questions = request_data.get('num_questions')
        difficulty = request_data.get('difficulty')
        
        # Normalize the topic
        clean_name = normalize_topic(topic)
        logger.info(f"Normalized topic: {clean_name}")

        # Initialize Gemini (synchronous version)
        llm = ChatGoogleGenerativeAI(
            model="gemini-1.5-flash",
            google_api_key=os.getenv("API_KEY"),
            temperature=0.9,
        )
        logger.info("Gemini LLM initialized")

        # Fetch past questions (synchronous version)
        past_questions = get_past_questions(user_id, clean_name)
        past_questions_str = "; ".join(past_questions) if past_questions else "None"
        logger.info(f"Past questions fetched: {past_questions}")

        # Dynamic prompt with randomization and history
        random_seed = random.randint(1, 1000)
        prompt = ChatPromptTemplate.from_messages([
            ("human", """
            Generate a quiz with the following details:
            - **Topic**: "{topic}"
            - **Number of Questions**: {num}
            - **Difficulty Level**: "{diff}"
            
            ### Instructions:
            1. Create an array of quiz questions in this JSON format: {format}
            2. Each question must include: serial_number (string), question, options (4 strings), correctAnswer, description.
            3. Use a creative twist (e.g., real-world scenarios, analogies) to add variety. Random seed: {seed}.
            4. Do NOT repeat these past questions: {past}.
            
            ### Guidelines:
            - Ensure content is accurate, clear, and matches the topic and difficulty.
            - If you can't generate the requested number, provide as many as possible.
            
            ### Fallback:
            - If the topic is inappropriate, return: [ {{ "ok": false }}, ["The requested topic is inappropriate or cannot be used."] ]
            
            Always return a two-element array: [ {{ "ok": bool }}, [questions or error] ].
            """)
        ])
        formatted_prompt = prompt.format(
            topic=clean_name,
            num=num_questions,
            diff=difficulty,
            format=output_parser.get_format_instructions(),
            seed=random_seed,
            past=past_questions_str
        )
        
        logger.info("Invoking LLM with prompt")
        # Synchronous invocation
        response = llm.invoke(formatted_prompt)
        logger.info(f"LLM response raw content: {response.content}")

        parsed_response = output_parser.parse(response.content)
        logger.info(f"Parsed response: {parsed_response}")

        if parsed_response["ok"]:
            if isinstance(parsed_response["data"][0], dict):
                logger.info("Returning valid quiz response")
                return parsed_response
            elif isinstance(parsed_response["data"][0], str):
                logger.error(f"Error in response data: {parsed_response['data'][0]}")
                raise Exception(parsed_response["data"][0])
            else:
                logger.error("Invalid response format")
                raise Exception("Invalid response format")
        else:
            logger.error("Quiz generation failed !!!!")
            return parsed_response
            
    except Exception as e:
        logger.error(f"Exception in generate_quiz 😓: {str(e)}")
        return {"ok": False, "data": [f"Error generating quiz: {str(e)}"]}