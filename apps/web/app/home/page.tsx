import Page from "@/components/layout/Page";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { MdOutlineQrCode2, MdOutlineViewList, MdOutlineAddCard, MdSend, MdSettings, MdEmail } from "react-icons/md";
import image1 from "@/public/images/carousel1.jpeg";
import image2 from "@/public/images/carousel2.jpeg";
import image3 from "@/public/images/carousel3.jpeg";
import Image from "next/image";

const HomePage = () => {
  return (
    <Page>
      <div className="flex min-h-full w-full flex-col gap-8 items-center">
        {/* Profile section */}
        <section className="bg-kimo-500 w-full px-4 h-40 flex flex-col py-4 sm:flex-row sm:py-0 sm:items-center justify-between gap-4">
          <div className="flex items-center gap-2">
            <Avatar size="2xl">
              <AvatarImage src="https://github.com/shadcn.png" />
              <AvatarFallback>CN</AvatarFallback>
            </Avatar>

            <div className="flex flex-col">
              <p className="font-bold text-xl text-white">Jhon Doe</p>
              <p className="text-sm text-white">+6281234567890</p>
            </div>
          </div>

          <div className="flex flex-col">
            <p className="font-bold text-xl text-white text-left sm:text-right">My Balance</p>
            <p className="text-2xl text-white text-left sm:text-right">Rp 100.000</p>
          </div>
        </section>

        {/* Menus */}
        <section className="bg-white px-4 sm:px-10 flex flex-col w-full">
          <div className="rounded-md p-4 grid grid-cols-3 gap-y-8 bg-kimo-50">
            <div className="flex flex-col items-center cursor-pointer hover:bg-white py-4 rounded-md transition-all duration-300">
              <MdOutlineAddCard size={28} className="text-kimo-500" />
              <p className="text-sm text-center">Top Up</p>
              <p className="text-xs text-gray-500 text-center">Top up your balance</p>
            </div>

            <div className="flex flex-col items-center cursor-pointer hover:bg-white py-4 rounded-md transition-all duration-300">
              <MdSend size={28} className="text-kimo-500" />
              <p className="text-sm text-center">Transfer</p>
              <p className="text-xs text-gray-500 text-center">Send money</p>
            </div>

            <div className="flex flex-col items-center cursor-pointer hover:bg-white py-4 rounded-md transition-all duration-300">
              <MdOutlineQrCode2 size={28} className="text-kimo-500" />
              <p className="text-sm text-center">QR Code</p>
              <p className="text-xs text-gray-500 text-center">Generate QR</p>
            </div>

            <div className="flex flex-col items-center cursor-pointer hover:bg-white py-4 rounded-md transition-all duration-300">
              <MdOutlineViewList size={28} className="text-kimo-500" />
              <p className="text-sm text-center">Transaction History</p>
              <p className="text-xs text-gray-500 text-center">See your history</p>
            </div>

            <div className="flex flex-col items-center cursor-pointer hover:bg-white py-4 rounded-md transition-all duration-300">
              <MdEmail size={28} className="text-kimo-500" />
              <p className="text-sm text-center">Inbox</p>
              <p className="text-xs text-gray-500 text-center">See recent notifications</p>
            </div>

            <div className="flex flex-col items-center cursor-pointer hover:bg-white py-4 rounded-md transition-all duration-300">
              <MdSettings size={28} className="text-kimo-500" />
              <p className="text-sm text-center">Settings</p>
              <p className="text-xs text-gray-500 text-center">Application settings</p>
            </div>
          </div>
        </section>

        {/* Promotion carousel */}
        <section className="flex w-full min-w-0 items-center gap-8 overflow-x-auto px-10">
          <div className="w-96 shrink-0 rounded-md overflow-hidden shadow-lg">
            <Image src={image1} alt="image-1" />
          </div>
          <div className="w-96 shrink-0 rounded-md overflow-hidden shadow-lg">
            <Image src={image2} alt="image-2" />
          </div>
          <div className="w-96 shrink-0 rounded-md overflow-hidden shadow-lg">
            <Image src={image3} alt="image-3" />
          </div>
        </section>

        {/* Latest 5 transactions */}
        <section className="w-full flex flex-col gap-8 px-10 border-t-2 bg-[#F7F7F7] py-4">
          <p className="font-medium text-left text-[#2C2C2C] text-lg">Latest Transactions</p>

          <div className="flex flex-col gap-4">
            <div className="flex items-center sm:w-80 justify-between">
              <div className="flex items-center gap-2">
                <Avatar size="lg">
                  <AvatarImage src="https://github.com/shadcn.png" />
                  <AvatarFallback>CN</AvatarFallback>
                </Avatar>

                <div className="flex flex-col">
                  <p>Jane Doe</p>
                  <p className="text-xs">15 Aug 2026, 17:20</p>
                </div>
              </div>

              <Badge variant="destructive">- Rp 50.000</Badge>
            </div>

            <div className="flex items-center sm:w-80 justify-between">
              <div className="flex items-center gap-2">
                <Avatar size="lg">
                  <AvatarImage src="https://github.com/shadcn.png" />
                  <AvatarFallback>CN</AvatarFallback>
                </Avatar>

                <div className="flex flex-col">
                  <p>John Smith</p>
                  <p className="text-xs">15 Aug 2026, 13:00</p>
                </div>
              </div>

              <Badge className="bg-green-50 text-green-700 dark:bg-green-950 dark:text-green-300">+ Rp 100.000</Badge>
            </div>

            <div className="flex items-center sm:w-80 justify-between">
              <div className="flex items-center gap-2">
                <Avatar size="lg">
                  <AvatarImage src="https://github.com/shadcn.png" />
                  <AvatarFallback>CN</AvatarFallback>
                </Avatar>

                <div className="flex flex-col">
                  <p>Jane Doe</p>
                  <p className="text-xs">15 Aug 2026, 17:20</p>
                </div>
              </div>

              <Badge variant="destructive">- Rp 50.000</Badge>
            </div>

            <div className="flex items-center sm:w-80 justify-between">
              <div className="flex items-center gap-2">
                <Avatar size="lg">
                  <AvatarImage src="https://github.com/shadcn.png" />
                  <AvatarFallback>CN</AvatarFallback>
                </Avatar>

                <div className="flex flex-col">
                  <p>Jane Doe</p>
                  <p className="text-xs">15 Aug 2026, 17:20</p>
                </div>
              </div>

              <Badge variant="destructive">- Rp 50.000</Badge>
            </div>

            <div className="flex items-center sm:w-80 justify-between">
              <div className="flex items-center gap-2">
                <Avatar size="lg">
                  <AvatarImage src="https://github.com/shadcn.png" />
                  <AvatarFallback>CN</AvatarFallback>
                </Avatar>

                <div className="flex flex-col">
                  <p>Jane Doe</p>
                  <p className="text-xs">15 Aug 2026, 17:20</p>
                </div>
              </div>

              <Badge variant="destructive">- Rp 50.000</Badge>
            </div>
          </div>
        </section>
      </div>
    </Page>
  );
};

export default HomePage;
