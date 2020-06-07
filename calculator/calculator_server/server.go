package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"net"

	"github.com/marprin/tutorial-go/calculator/calculatorpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct{}

func (*server) Sum(ctx context.Context, req *calculatorpb.SumRequest) (*calculatorpb.SumResponse, error) {
	fmt.Printf("Receive Sum RPC: %v", req)
	firstNumber := req.GetFirstNumber()
	secondNumber := req.GetSecondNumber()

	sum := firstNumber + secondNumber
	res := &calculatorpb.SumResponse{
		SumResult: sum,
	}

	return res, nil
}

func (*server) PrimeNumberDecomposition(req *calculatorpb.PrimeNumberDecompositionRequest, stream calculatorpb.CalculatorService_PrimeNumberDecompositionServer) error {
	fmt.Printf("Received PrimeNumberDecomposition RPC %v\n", req)

	number := req.GetNumber()
	divisor := int64(2)

	for number > 1 {
		if number%divisor == 0 {
			stream.Send(&calculatorpb.PrimeNumberDecompositionResponse{
				PrimeFactor: divisor,
			})

			number = number / divisor
		} else {
			divisor++
			fmt.Printf("Divisor has increased to %v", divisor)
		}
	}
	return nil
}

func (*server) ComputeAverage(stream calculatorpb.CalculatorService_ComputeAverageServer) error {
	fmt.Printf("Received ComputeAverage RPC \n")

	sum := int32(0)
	count := 0

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			avg := float64(sum) / float64(count)
			return stream.SendAndClose(&calculatorpb.ComputeAverageResponse{
				Average: avg,
			})
		}
		if err != nil {
			log.Fatalf("err while reading client stream: %v\n", err)
		}

		sum += req.GetNumber()
		count++
	}
}
func (*server) FindMaximum(stream calculatorpb.CalculatorService_FindMaximumServer) error {
	fmt.Printf("Received FindMaximum RPC \n")
	maximum := int32(0)

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("Error while reading client stream: %v\n", err)
		}

		number := req.GetNumber()
		if number > maximum {
			maximum = number
			sendErr := stream.Send(&calculatorpb.FindMaximumResponse{
				Maximum: maximum,
			})

			if sendErr != nil {
				log.Fatalf("Error while sending data to client: %v\n", sendErr)
				return sendErr
			}
		}
	}
}

func (*server) SquareRoot(ctx context.Context, req *calculatorpb.SquareRootRequest) (*calculatorpb.SquareRootResponse, error) {
	fmt.Printf("Received SquareRoot RPC \n")

	number := req.GetNumber()
	if number < 0 {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Received a negative number: %v", number),
		)
	}
	return &calculatorpb.SquareRootResponse{
		NumberRoot: math.Sqrt(float64(number)),
	}, nil
}

func main() {
	fmt.Println("Calculator Server")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v\n", err)
	}

	s := grpc.NewServer()
	calculatorpb.RegisterCalculatorServiceServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v\n", err)
	}

}
