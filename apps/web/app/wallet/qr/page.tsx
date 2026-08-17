import Page from "@/components/layout/Page";
import { Button } from "@/components/ui/button";
import QRCode from "react-qr-code";
import { MdArrowBack, MdOutlineFileUpload, MdOutlineContentCopy } from "react-icons/md";
import Link from "next/link";

const QRISPage = () => {
  return (
    <Page>
      <div className="bg-kimo-500 flex min-h-full w-full flex-col items-center justify-between py-10 px-4">
        <div className="relative w-full flex justify-center">
          <p className="text-white text-lg font-medium">Request from QRIS</p>

          <Link
            href={"/home"}
            className="absolute top-0.5 left-0 sm:left-10 flex items-center gap-2 font-medium text-white"
          >
            <MdArrowBack size={24} />
            <p className="hidden sm:flex">Back</p>
          </Link>
        </div>

        <div className="flex flex-col gap-2">
          <div className="h-auto m-0 max-w-72 bg-white p-4 rounded-lg">
            <QRCode
              size={256}
              style={{ height: "auto", maxWidth: "100%", width: "100%" }}
              value={
                "00020101021226670015ID.SINGAPAY.WWW01189360783904122600420152886407023049170303UME51440014ID.CO.QRIS.WWW0215ID190900099006003UME6221051017389001790703CO153033605405110005502035702306015JAKARTASELATAN520458125914PAYCRED_ACQ_015802ID610458126304423E"
              }
              viewBox={`0 0 256 256`}
            />
          </div>

          <Button>
            <MdOutlineContentCopy size={20} />
            Copy code
          </Button>
        </div>

        <Button className="w-full">
          <MdOutlineFileUpload size={20} />
          SHARE QR CODE
        </Button>
      </div>
    </Page>
  );
};

export default QRISPage;
