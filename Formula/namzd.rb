class Namzd < Formula
  desc "Quickly find files by name or extension"
  homepage "https://github.com/bengarrett/namzd"
  url "https://github.com/bengarrett/namzd/archive/refs/tags/v1.3.3.tar.gz"
  sha256 "3135689c07b5d60219cf07a3105fb3af00ee80f110100dfb2c3234eccc9bbfe4"
  version "1.3.3"
  license "GPL-3.0-only"

  @commit = "0dce4d208550be5588ad27076fefae03f4710e8f"
  @build_date = "2026-10-28T14:44:42+11:00"

  livecheck do
    url :stable
    strategy :github_latest
  end

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w -X main.version=#{version} -X main.commit=#{self.class.instance_variable_get('@commit')} -X main.date=#{self.class.instance_variable_get('@build_date')}")
  end

  test do
    assert_match "namzd", shell_output("#{bin}/namzd --version")
  end
end
