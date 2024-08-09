from github import Github
import re
from croniter import croniter
from datetime import datetime, timedelta
import getpass

GITHUB_URL = 'https://github.example.com/api/v3'
ORGANIZATION = 'ORG_NAME_HERE'

# 특정 레포지토리 이름을 추가하거나 빈 리스트로 둠
# 빈 리스트로 둘 경우, organization에 속한 전체 레포지터리를 탐색합니다.
REPO_NAMES = []

def convert_cron_to_kst(cron_expr):
    """
    Convert a cron expression from UTC to KST.
    Assumes cron_expr is a valid cron expression.
    """
    base_time = datetime.now()
    iter = croniter(cron_expr, base_time)
    next_run = iter.get_next(datetime)
    
    # Convert UTC to KST (UTC + 9 hours)
    kst_time = next_run + timedelta(hours=9)
    return kst_time.strftime('%Y-%m-%d %H:%M:%S')

def search_workflow_files(g, org, repo_names):
    if repo_names:
        # Search specified repositories in the organization
        repos = [g.get_repo(f"{org}/{repo_name}") for repo_name in repo_names]
    else:
        # Search all repositories in the organization
        repos = g.get_organization(org).get_repos()
    
    for idx, repo in enumerate(repos, start=1):
        # Print current repo being searched with index
        print(f"Searching {repo.full_name} ({idx}/{repos.totalCount if not repo_names else len(repos)}) ...")

        try:
            # List repository contents (files and directories)
            contents = repo.get_contents("")
        except Exception as e:
            print(f"Error accessing {repo.full_name}: {e}")
            continue
        
        while contents:
            file_content = contents.pop(0)
            if file_content.type == "dir":
                contents.extend(repo.get_contents(file_content.path))
            else:
                # Check if the file matches the search criteria
                if file_content.path.startswith('.github/workflows/') and file_content.path.endswith(('.yaml', '.yml')):
                    # Read file content to search for 'schedule:'
                    file_data = repo.get_contents(file_content.path).decoded_content.decode('utf-8')
                    if 'schedule:' in file_data:
                        # Print repository name and file path
                        print(f"{repo.full_name} - Found 'schedule:' in {file_content.path}")

                        # Find and convert cron expressions
                        cron_expressions = re.findall(r"cron:\s*'([^']+)'", file_data)
                        for cron_expr in cron_expressions:
                            kst_time = convert_cron_to_kst(cron_expr)
                            print(f"  Original cron: {cron_expr}, KST Time: {kst_time}")

def main():
    # Get Access Token from user input
    access_token = getpass.getpass("Enter your GitHub Access Token: ")

    # Authenticate to GitHub Enterprise Server
    g = Github(base_url=GITHUB_URL, login_or_token=access_token)

    # Search workflow files
    print(f"Searching for 'schedule:' in .github/workflows/*.yaml or .github/workflows/*.yml files...")
    search_workflow_files(g, ORGANIZATION, REPO_NAMES)

if __name__ == "__main__":
    main()