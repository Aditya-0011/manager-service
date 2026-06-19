package faults

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrInternal             = status.Error(codes.Internal, "Internal server error")
	ErrInternalCreate       = status.Error(codes.Internal, "Error creating resource")
	ErrInternalUpdate       = status.Error(codes.Internal, "Error updating resource")
	ErrInternalDelete       = status.Error(codes.Internal, "Error deleting resource")
	ErrExperienceNotFound   = status.Error(codes.NotFound, "Experience not found")
	ErrProjectNotFound      = status.Error(codes.NotFound, "Project not found")
	ErrTechnologyNotFound   = status.Error(codes.NotFound, "Technology not found")
	ErrMessageNotFound      = status.Error(codes.NotFound, "Message not found")
	ErrExperiencesNotFound  = status.Error(codes.NotFound, "Experiences not found")
	ErrProjectsNotFound     = status.Error(codes.NotFound, "Projects not found")
	ErrTechnologiesNotFound = status.Error(codes.NotFound, "Technologies not found")
	ErrMessagesNotFound     = status.Error(codes.NotFound, "Messages not found")
	ErrUserNotFound         = status.Error(codes.NotFound, "User not found")
	ErrMessageLimit         = status.Error(codes.Aborted, "You have reached the limit of messages")
)
