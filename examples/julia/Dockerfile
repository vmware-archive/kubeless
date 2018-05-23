FROM knightxun/julia
RUN mkdir -p /kubeless

ENV FUNC_HANDLER=foo \
    MOD_NAME=hello
COPY hello.jl /kubeless
COPY REQUIRE  /root/.julia/v0.6/REQUIRE
RUN julia -e "Pkg.resolve()"
CMD ["python3", "/kubeless-julia.py"]

