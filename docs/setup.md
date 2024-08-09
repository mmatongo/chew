# Setting up Google Cloud Services for Speech-to-Text

1. **Create a Google Cloud Project**
   - Go to the [Google Cloud Console](https://console.cloud.google.com/)
   - Click on the project dropdown and select "New Project"
   - Enter a project name and click "Create"

2. **Enable the Cloud Speech-to-Text API**
   - In the Google Cloud Console, under "Quick access" go to "APIs & Services"
   - Click on "+ ENABLE APIS AND SERVICES"
   - Search for "Cloud Speech-to-Text API" and select it
   - Click "Enable"

3. **Create a Service Account**
   - In the Google Cloud Console, go to "IAM & Admin" > "Service Accounts" (or use [this link](https://console.cloud.google.com/iam-admin/serviceaccounts))
   - Click "Create Service Account"
   - Enter a name for the service account and click "Create"
   - For the role, choose "Project" > "Owner" (or a more restrictive role if preferred)
   - Click "Continue" and then "Done"

4. **Generate a Key for the Service Account**
   - In the Service Accounts list, find the account you just created
   - Click on the three dots menu (⋮) and select "Manage keys"
   - Click "Add Key" > "Create new key"
   - Choose "JSON" as the key type and click "Create"
   - The key file will be downloaded to your computer

5. **Set the GOOGLE_APPLICATION_CREDENTIALS Environment Variable**
   - On Linux or macOS:
     ```
     export GOOGLE_APPLICATION_CREDENTIALS="/path/to/your/service-account-key.json"
     ```
   - If you want to set the environment variable permanently, you can add it to your shell profile (e.g., `~/.bashrc`, `~/.zshrc`, etc.)

   You can optionally set the environment variable in your code as well:
   ```python
    import os

    os.environ["GOOGLE_APPLICATION_CREDENTIALS"] = "/path/to/your/service-account-key.json"
    ```

    ```go
    import "os"

    os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/path/to/your/service-account-key.json")
    ```

    ```ruby
    ENV["GOOGLE_APPLICATION_CREDENTIALS"] = "/path/to/your/service-account-key.json"
    ```
