import os 
from functools import wraps
from flask import current_app, request, jsonify

API_KEY_ENVIRONMENT_VARIABLE = "MANAGER_API_KEY"

def require_authentication(view_function):
    @wraps(view_function)
    def decorated_function(*args, **kwargs):
        expected_api_key = os.environ.get(API_KEY_ENVIRONMENT_VARIABLE)

        provided_api_key = request.headers.get("X-API-KEY")

        if not expected_api_key:
            current_app.logger.error("Server Configueration Error: API KEY not in environment")
            return jsonify({"error": "Server Configuration Error"}), 500
        
        if provided_api_key and provided_api_key == expected_api_key:
            return view_function(*args, **kwargs)
        
        current_app.logger.warning(f"Unauthorized access attempt from {request.remote_addr}")
        return jsonify({"error": "Unauthorized"}), 401
    
    return decorated_function