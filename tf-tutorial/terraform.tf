terraform {
  backend "gcs" {
    bucket = "tp-learning-bucket-tfstate"
    prefix = "terraform/state"
  }
}
