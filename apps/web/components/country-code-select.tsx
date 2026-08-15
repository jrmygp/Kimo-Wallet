"use client";

import * as React from "react";
import { all } from "country-codes-list";
import * as Flags from "country-flag-icons/react/3x2";
import { ChevronDownIcon, SearchIcon } from "lucide-react";
import { InputGroup, InputGroupAddon, InputGroupInput } from "@/components/ui/input-group";
import { cn } from "@/lib/utils";
import { Sheet, SheetContent, SheetClose, SheetHeader, SheetTitle, SheetTrigger } from "@/components/ui/sheet";
import { MdOutlineClose } from "react-icons/md";

type FlagComponent = (typeof Flags)["US"];

const countries = all().sort((a, b) => a.countryNameEn.localeCompare(b.countryNameEn));

function CountryFlag({ countryCode, className }: { countryCode: string; className?: string }) {
  const Flag = (Flags as Record<string, FlagComponent | undefined>)[countryCode];

  if (!Flag) {
    return <span aria-hidden className={cn("shrink-0 rounded-[3px] bg-muted", className)} />;
  }

  return <Flag aria-hidden className={cn("shrink-0 rounded-[3px] object-cover", className)} />;
}

export interface CountryCodeSelectProps {
  /** ISO 3166-1 alpha-2 country code, e.g. "ID" */
  value: string;
  onChange: (countryCode: string) => void;
  disabled?: boolean;
  placeholder?: string;
  className?: string;
}

export function CountryCodeSelect({
  value,
  onChange,
  disabled,
  placeholder = "Select country",
  className,
}: CountryCodeSelectProps) {
  const [open, setOpen] = React.useState(false);
  const [query, setQuery] = React.useState("");

  const selected = React.useMemo(() => countries.find((country) => country.countryCode === value), [value]);

  const filtered = React.useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return countries;
    return countries.filter(
      (country) =>
        country.countryNameEn.toLowerCase().includes(q) ||
        country.countryCallingCode.includes(q) ||
        country.countryCode.toLowerCase().includes(q),
    );
  }, [query]);

  function handleOpenChange(next: boolean) {
    setOpen(next);
    if (!next) setQuery("");
  }

  function handleSelect(countryCode: string) {
    onChange(countryCode);
    handleOpenChange(false);
  }

  return (
    <Sheet open={open} onOpenChange={handleOpenChange}>
      <SheetTrigger
        disabled={disabled}
        className={cn(
          "flex h-12 items-center gap-1.5 rounded-sm bg-background px-2.5 text-sm shadow-xs outline-none transition-colors hover:bg-muted disabled:pointer-events-none disabled:opacity-50 focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 dark:border-input dark:bg-input/30",
          className,
        )}
      >
        {selected ? (
          <>
            <CountryFlag countryCode={selected.countryCode} className="h-3.5 w-5 border" />
            <span className="font-medium">+{selected.countryCallingCode}</span>
          </>
        ) : (
          <span>{placeholder}</span>
        )}
        <ChevronDownIcon className="size-3.5 text-muted-foreground" />
      </SheetTrigger>

      <SheetContent
        side="bottom"
        showCloseButton={false}
        className="data-[side=bottom]:h-[80vh] gap-0 p-0 rounded-t-2xl"
      >
        <SheetHeader className="gap-3 border-b border-border pb-4">
          <div className="flex items-center justify-between">
            <SheetTitle>Select country code</SheetTitle>
            <SheetClose>
              <MdOutlineClose size={20} className="cursor-pointer"/>
            </SheetClose>
          </div>

          <InputGroup className="border">
            <InputGroupAddon>
              <SearchIcon />
            </InputGroupAddon>

            <InputGroupInput
              autoFocus
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              placeholder="Search country or code"
            />
          </InputGroup>
        </SheetHeader>

        <div role="listbox" aria-label="Countries" className="min-h-0 flex-1 overflow-y-auto px-2 py-2">
          {filtered.length === 0 ? (
            <p className="px-2 py-6 text-center text-sm text-muted-foreground">No country found.</p>
          ) : (
            filtered.map((country) => (
              <button
                key={country.countryCode}
                type="button"
                role="option"
                aria-selected={country.countryCode === value}
                onClick={() => handleSelect(country.countryCode)}
                className={cn(
                  "flex w-full items-center gap-3 rounded-sm px-2.5 py-2.5 text-left text-sm hover:bg-muted",
                  country.countryCode === value && "bg-muted",
                )}
              >
                <CountryFlag countryCode={country.countryCode} className="h-4 w-6" />
                <span className="flex-1 truncate">{country.countryNameEn}</span>
                <span className="text-muted-foreground">+{country.countryCallingCode}</span>
              </button>
            ))
          )}
        </div>
      </SheetContent>
    </Sheet>
  );
}
