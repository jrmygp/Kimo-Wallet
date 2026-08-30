"use client";

import { use } from "react";
import Page from "@/components/layout/Page";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { useSearchUserQuery } from "@/features/wallet/hooks/use-search-user-query";

const TransferPage = ({ params }: { params: Promise<{ userId: string }> }) => {
  const { userId } = use(params);
  const { data, isFetching, isError, error } = useSearchUserQuery(userId);

  return (
    <Page>
      <div className="flex min-h-full w-full flex-col gap-8 items-center relative bg-[#F5F5F5]">
        <section className="bg-kimo-500 w-full px-4 h-40 flex py-4 items-center justify-center">
          <p className="text-2xl font-medium text-white">Send to Friend</p>
        </section>

        <div className="absolute top-30 w-[80%] flex flex-col bg-white rounded-md p-4">
          <div className="flex items-center gap-2">
            <Avatar size="2xl">
              <AvatarImage src={data?.profilePicture || "https://github.com/shadcn.png"} />
              <AvatarFallback>{data?.fullName.slice(0, 2).toUpperCase() ?? "CN"}</AvatarFallback>
            </Avatar>

            <div>
              {data && <p className="text-lg font-medium">{data.fullName}</p>}
              {data && <p className="text-sm">{data.phoneNumber}</p>}
            </div>
          </div>

          {isFetching && <p className="text-sm text-muted-foreground">Loading recipient...</p>}
          {isError && (
            <p role="alert" className="text-sm text-destructive">
              {error.message}
            </p>
          )}
        </div>
      </div>
    </Page>
  );
};

export default TransferPage;
