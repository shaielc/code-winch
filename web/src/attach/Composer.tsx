import {
  useId,
  useRef,
  useState,
  type FormEvent,
  type KeyboardEvent,
} from "react";

interface ComposerProps {
  onSubmit: (message: string) => void;
}

const validationMessage = "Enter a message before sending.";

export function Composer({ onSubmit }: ComposerProps) {
  const [draft, setDraft] = useState("");
  const [validationError, setValidationError] = useState(false);
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const textareaId = useId();
  const validationId = `${textareaId}-validation`;

  function submit(event: FormEvent) {
    event.preventDefault();

    if (draft.trim().length === 0) {
      setValidationError(true);
      return;
    }

    onSubmit(draft);
    setDraft("");
    setValidationError(false);
    textareaRef.current?.focus();
  }

  function handleKeyDown(event: KeyboardEvent<HTMLTextAreaElement>) {
    if (event.key === "Enter" && !event.shiftKey) {
      event.preventDefault();
      submit(event);
    }
  }

  return (
    <form className="composer" onSubmit={submit}>
      <label htmlFor={textareaId}>Message</label>
      <textarea
        id={textareaId}
        ref={textareaRef}
        value={draft}
        aria-describedby={validationError ? validationId : undefined}
        onChange={(event) => {
          const value = event.target.value;
          setDraft(value);
          if (value.trim().length > 0) setValidationError(false);
        }}
        onKeyDown={handleKeyDown}
        rows={4}
      />
      {validationError && (
        <p id={validationId} className="validation" role="alert">
          {validationMessage}
        </p>
      )}
      <button type="submit">Send message</button>
    </form>
  );
}
