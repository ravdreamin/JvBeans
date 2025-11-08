"use client";

import { useState, useRef, useEffect } from "react";

interface InlineNameInputProps {
  initial?: string;
  placeholder?: string;
  onSubmit: (value: string) => Promise<void> | void;
  onCancel: () => void;
  validate?: (value: string) => string | null;
  autoFocus?: boolean;
}

export default function InlineNameInput({
  initial = "",
  placeholder = "Enter name...",
  onSubmit,
  onCancel,
  validate,
  autoFocus = true,
}: InlineNameInputProps) {
  const [value, setValue] = useState(initial);
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (autoFocus && inputRef.current) {
      inputRef.current.focus();
      inputRef.current.select();
    }
  }, [autoFocus]);

  const handleSubmit = async () => {
    const trimmed = value.trim();
    if (!trimmed) {
      setError("Name cannot be empty");
      return;
    }

    if (validate) {
      const validationError = validate(trimmed);
      if (validationError) {
        setError(validationError);
        return;
      }
    }

    setIsSubmitting(true);
    setError(null);

    try {
      await onSubmit(trimmed);
    } catch (err: any) {
      setError(err.message || "Failed to save");
      setIsSubmitting(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") {
      e.preventDefault();
      handleSubmit();
    } else if (e.key === "Escape") {
      e.preventDefault();
      onCancel();
    }
  };

  return (
    <div className="flex flex-col gap-1 py-1">
      <input
        ref={inputRef}
        type="text"
        value={value}
        onChange={(e) => {
          setValue(e.target.value);
          setError(null);
        }}
        onKeyDown={handleKeyDown}
        onBlur={onCancel}
        placeholder={placeholder}
        disabled={isSubmitting}
        className="w-full px-2 py-1 text-sm bg-black border border-primaryBlue rounded focus:outline-none focus:ring-1 focus:ring-primaryBlue text-white disabled:opacity-50"
      />
      {error && <span className="text-xs text-red-400 px-2">{error}</span>}
    </div>
  );
}
