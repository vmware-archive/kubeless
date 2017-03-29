# Obtains the latest Kubeless release published 
class Kubelessfunction
        def self.run(request)
		require "net/https"
		require "uri"
		require "json"

		# Fetch release info
		uri = URI.parse("https://api.github.com/repos/bitnami/kubeless/releases")
		http = Net::HTTP.new(uri.host, uri.port)
		request = Net::HTTP::Get.new(uri.request_uri)
		http.use_ssl = true
		http.verify_mode = OpenSSL::SSL::VERIFY_PEER
		response = http.request(request)

		# Parse response
		output = JSON.parse(response.body)

		# Create a Hash for output 
		output_hash = { :version => output[0]['name'] }

		# Print the stuff (JSON)
		puts JSON.pretty_generate(output_hash)
        end
end
