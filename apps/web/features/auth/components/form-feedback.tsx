import { CircleAlertIcon } from "lucide-react";
import { FieldDescription } from "@/components/ui/field";
import { cn } from "@/lib/utils";

// Always rendered with a reserved line height, so showing an error never shifts the
// layout. Shows `hint` in muted text when there is no error.
export function FieldMessage({ id, error, hint }: { id: string; error?: string; hint?: string }) {
  return (
    <FieldDescription
      id={id}
      className={cn("flex min-h-5 items-center gap-1.5", error ? "text-destructive" : "text-muted-foreground")}
    >
      {error ? (
        <>
          <CircleAlertIcon className="size-4 shrink-0" aria-hidden />
          {error}
        </>
      ) : (
        hint
      )}
    </FieldDescription>
  );
}

export function FormAlert({ message }: { message: string }) {
  return (
    <div role="alert" className="mt-6 flex items-start gap-2 rounded-xl bg-destructive/10 px-3 py-3 text-sm text-destructive">
      <CircleAlertIcon className="mt-0.5 size-4 shrink-0" aria-hidden />
      <p>{message}</p>
    </div>
  );
}
