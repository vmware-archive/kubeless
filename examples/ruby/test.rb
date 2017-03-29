class Kubelessfunction
        def self.run(request)
          puts request.env
        end
end
