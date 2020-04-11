
resource "aws_ecs_cluster" "MaxEdgeECS" {
    name = "MaxEdge-POC-Cluster-Dev"
    capacity_providers = ["FARGATE"]
    
  
}