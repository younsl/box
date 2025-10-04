#===============================================================================
# Lambda Function
#===============================================================================
output "lambda_function_arn" {
  description = "The ARN of the Lambda Function"
  value       = module.lambda.lambda_function_arn
}

output "lambda_function_name" {
  description = "The name of the Lambda Function"
  value       = module.lambda.lambda_function_name
}
