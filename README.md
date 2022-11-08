# Pipelines as Code with Go, Dagger, and Terraform

Source code to accompany the YouTube mini series where I investigate dagger.io for building CI/CD
pipelines for Go applications.

## Source Code Orginaization
The main branch contains the latest code with the resultant code for each episode in
their own branches. If you would like to follow along with an episode, check out the
previous episodes branch. With the exception of part1, for that you can just use
an empty repo.

## Part 1 (branch part1) - Building a Basic Pipeline

In Part 1 of the series we look at the basics on how Dagger works and build
a pipeline that compiles our Go code, creates a Docker image and pushes
it to Docker Hub.

[https://allthings.tube/dagger1](https://allthings.tube/dagger1)

## Part 2 (branch part2) - Deploying Applications

In Part 2 of the series we investigate how to use Dagger with the
Terraform CDK to deploy our application only using Go.

[https://allthings.tube/dagger2](https://allthings.tube/dagger2)
