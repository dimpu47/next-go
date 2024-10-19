import React from "react";
import { render, screen, waitFor } from "@testing-library/react";
import axios from "axios";
import UserInterface from "./UserInterface"; 
import '@testing-library/jest-dom'; 

jest.mock("axios");
const mockedAxios = axios as jest.Mocked<typeof axios>;

describe("UserInterface Component", () => {
  const backendName = "go"; 

  // Mock console.error to suppress error messages in tests
  const originalConsoleError = console.error;

  beforeEach(() => {
    // Clear previous mock calls
    jest.clearAllMocks();
    console.error = jest.fn(); // Suppress console.error
  });

  afterEach(() => {
    console.error = originalConsoleError; // Restore original console.error after each test
  });

  it("renders without crashing", () => {
    render(<UserInterface backendName={backendName} />);
    expect(screen.getByText("Go Backend")).toBeInTheDocument();
  });
});
