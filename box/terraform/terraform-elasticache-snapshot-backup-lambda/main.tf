provider "aws" {
  region = "ap-northeast-2"

  # Make it faster by skipping something
  skip_metadata_api_check     = true
  skip_region_validation      = true
  skip_credentials_validation = true
  skip_requesting_account_id  = true
}

#===============================================================================
# Local variables
#===============================================================================
locals {
  name = "elasticache-snapshot-backup-lambda"

  tags = {
    Name        = local.name
    ManagedBy   = "terraform"
  }
}

#===============================================================================
# Lambda Function
#===============================================================================
module "lambda" {
  source  = "terraform-aws-modules/lambda/aws"
  version = "7.20.2"

  create_role  = false

  function_name = local.name
  description   = "Lambda Function for ${local.name}"

  handler       = "index.lambda_handler"
  runtime       = "python3.13"
  architectures = ["arm64"]

  memory_size = 128
  timeout     = 600

  create_package = true
  source_path    = "./src/index.py"


  create_current_version_allowed_triggers = false
  allowed_triggers = {
    EventBridge = {
      principal  = "events.amazonaws.com"
      source_arn = aws_cloudwatch_event_rule.schedule.arn
    }
  }

  environment_variables = {
    TZ = "Asia/Seoul"
  }

  lambda_role = "<LAMBDA_ROLE_ARN>"

  #===============================================================================
  # Additional inline policy
  #===============================================================================
  attach_policy_statements = false
  policy_statements        = {}

  #===============================================================================
  # Networking
  #===============================================================================
  vpc_subnet_ids         = ["<SUBNET_ID_1>", "<SUBNET_ID_2>"]
  vpc_security_group_ids = ["<SECURITY_GROUP_ID>"]
  attach_network_policy  = true

  tags = local.tags
}

#===============================================================================
# EventBridge Rule for Scheduled Trigger
#===============================================================================
resource "aws_cloudwatch_event_rule" "schedule" {
  name                = "${local.name}-schedule"
  description         = "Trigger ${local.name} every day at 00:10 KST"
  schedule_expression = "cron(10 15 * * ? *)" # 15:10 UTC = 00:10 KST (UTC+9)
  state               = "ENABLED"

  tags = local.tags
}

resource "aws_cloudwatch_event_target" "lambda" {
  rule      = aws_cloudwatch_event_rule.schedule.name
  target_id = "${local.name}-target"
  arn       = module.lambda.lambda_function_arn
}
