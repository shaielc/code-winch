import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, expect, test, vi } from "vitest";
import { Composer } from "./Composer";

afterEach(cleanup);

test("submits with Enter, then clears and focuses the composer", () => {
  const onSubmit = vi.fn();
  render(<Composer onSubmit={onSubmit} />);
  const composer = screen.getByRole("textbox", { name: "Message" });

  fireEvent.change(composer, { target: { value: "hello" } });
  fireEvent.keyDown(composer, { key: "Enter" });

  expect(onSubmit).toHaveBeenCalledWith("hello");
  expect(composer).toHaveValue("");
  expect(composer).toHaveFocus();
});

test("the button submits the exact non-empty payload", () => {
  const onSubmit = vi.fn();
  render(<Composer onSubmit={onSubmit} />);
  const composer = screen.getByRole("textbox", { name: "Message" });

  fireEvent.change(composer, { target: { value: "  keep me  " } });
  fireEvent.click(screen.getByRole("button", { name: "Send message" }));

  expect(onSubmit).toHaveBeenCalledWith("  keep me  ");
  expect(composer).toHaveValue("");
  expect(composer).toHaveFocus();
});

test("Shift+Enter does not submit and embedded whitespace is preserved", () => {
  const onSubmit = vi.fn();
  render(<Composer onSubmit={onSubmit} />);
  const composer = screen.getByRole("textbox", { name: "Message" });

  fireEvent.change(composer, { target: { value: "first" } });
  fireEvent.keyDown(composer, { key: "Enter", shiftKey: true });
  expect(onSubmit).not.toHaveBeenCalled();

  fireEvent.change(composer, { target: { value: " first\nsecond " } });
  fireEvent.keyDown(composer, { key: "Enter" });
  expect(onSubmit).toHaveBeenCalledWith(" first\nsecond ");
});

test("refuses blank input visibly without clearing or moving focus", () => {
  const onSubmit = vi.fn();
  render(<Composer onSubmit={onSubmit} />);
  const composer = screen.getByRole("textbox", { name: "Message" });
  composer.focus();

  fireEvent.change(composer, { target: { value: "   \n" } });
  fireEvent.keyDown(composer, { key: "Enter" });

  const alert = screen.getByRole("alert");
  expect(alert).toHaveTextContent("Enter a message before sending.");
  expect(composer).toHaveAttribute("aria-describedby", alert.id);
  expect(composer).toHaveValue("   \n");
  expect(composer).toHaveFocus();
  expect(onSubmit).not.toHaveBeenCalled();
});

test("clears the validation alert when the draft becomes non-whitespace", () => {
  render(<Composer onSubmit={vi.fn()} />);
  const composer = screen.getByRole("textbox", { name: "Message" });

  fireEvent.keyDown(composer, { key: "Enter" });
  expect(screen.getByRole("alert")).toBeInTheDocument();

  fireEvent.change(composer, { target: { value: "x" } });
  expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  expect(composer).not.toHaveAttribute("aria-describedby");
});
