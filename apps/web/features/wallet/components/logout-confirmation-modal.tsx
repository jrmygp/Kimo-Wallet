import { LogOutIcon } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

const LogoutConfirmationModal = ({ onLogout }: { onLogout: () => void }) => {
  return (
    <DialogContent className="sm:max-w-sm">
      <DialogHeader className="items-center text-center">
        <div className="flex size-12 items-center justify-center rounded-full bg-kimo-50 text-primary">
          <LogOutIcon className="size-6" aria-hidden />
        </div>
        <DialogTitle>Log out of Kimo Wallet?</DialogTitle>
        <DialogDescription>
          You&apos;ll need to log in again to view your balance and send money.
        </DialogDescription>
      </DialogHeader>

      <DialogFooter>
        <DialogClose render={<Button variant="outline">Cancel</Button>} />
        <Button type="button" variant="primary" onClick={onLogout}>
          Log out
        </Button>
      </DialogFooter>
    </DialogContent>
  );
};

export default LogoutConfirmationModal;
