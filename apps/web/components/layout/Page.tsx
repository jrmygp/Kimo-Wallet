import React, { ReactNode } from "react";

interface PageProps {
  children: ReactNode;
}

const Page = ({ children }: PageProps) => {
  return (
    <div className="flex items-center justify-center h-screen">
      <div className="mx-auto h-screen w-full max-w-3xl overflow-auto sm:rounded-2xl rounded-none shadow-lg">
        {children}
      </div>
    </div>
  );
};

export default Page;
