import subprocess
import logging
import sys
import os
from flask import Flask, jsonify, request
from auth_middleware import require_authentication

application = Flask(__name__)

log_directory = "../logs"
os.makedirs(log_directory, exist_ok=True)
logging.basicConfig(
    filename=os.path.join(log_directory, "python_service.log"),
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)

@application.route("/libraries", methods=["GET"])
@require_authentication
def get_installed_libraries():
    try:
        process_result = subprocess.run(
            [sys.executable, '-m', 'pip', 'list', '--format=json'],
            capture_output=True,
            text=True
        )

        if process_result.returncode != 0:
            logging.error(f"Failed to list packages: {process_result.stderr}")
            return jsonify({"error": "Failed to retrieve packages"}), 500
        
        import json
        packages_list = json.loads(process_result.stdout)
        logging.info(f"Retrieved {len(packages_list)} packages")
        return jsonify(packages_list), 200
    
    except Exception as GeneralException:
        logging.exception("An unexpected error occured while fetching libraries...")
        return jsonify({"error": str(GeneralException)}), 500

@application.route("/libraries", methods=["POST"])
@require_authentication
def install_library():
    request_data = request.get_json()
    package_name = request_data.get("name")

    if not package_name:
        return jsonify({"error": "Package name is required"}), 400
    
    try:
        logging.info(f"Installing package: {package_name}")
        process_result = subprocess.run(
            [sys.executable, "-m", "pip", "install", package_name],
            capture_output=True,
            text=True
        )

        if process_result.returncode != 0:
            logging.info(f"Installed: {package_name}")
            return jsonify({"message": f"Installed {package_name}", "output": process_result.stdout}), 200
        else:
            logging.error(f"Installation failed for {package_name}: {process_result.stderr}")
            return jsonify({"error": "Installation failed", "details": process_result.stderr}), 400
    
    except Exception as GeneralException:
        logging.exception(f"Exception during installation of {package_name}")
        return jsonify({"error": str(GeneralException)}), 500

@application.route("/libraries/<package_name>", methods=["DELETE"])
@require_authentication
def delete_library(package_name):
    try:
        logging.info(f"Attempting to uninstall package: {package_name}")
        process_result = subprocess.run(
            [sys.executable, '-m', 'pip', 'uninstall', '-y', package_name],
            capture_output=True,
            text=True
        )

        if process_result.returncode == 0:
            logging.info(f"Successfully uninstalled {package_name}")
            return jsonify({"message": f"Successfully uninstalled {package_name}", "output": process_result.stdout}), 200
        else:
            logging.error(f"Uninstallation failed for {package_name}: {process_result.stderr}")
            return jsonify({"error": "Uninstallation failed", "details": process_result.stderr}), 400

    except Exception as GeneralException:
        logging.exception(f"Exception during uninstallation of {package_name}")
        return jsonify({"error": str(GeneralException)}), 500

if __name__ == "__main__":
    print("Starting Python Server on http://127.0.0.1:5000...")
    application.run(host="127.0.0.1", port=5000, debug=False)