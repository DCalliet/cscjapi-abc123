# Programming Test


## Scenario:


You are given the task of building a job processing system. For the purpose of this exercise, the system will perform 2
core tasks:


-  Listen for HTTP requests, and perform the following actions:

    - Allow a caller to request a JSON object listing the current jobs within the queue

    - Allow the jobs to be filtered by a single “status” parameter, if one is supplied (otherwise return all jobs)

    - Allow a caller to add jobs to the queue by submitting POST data to the API


- A background thread (or additional service) that processes jobs within the queue


- A “processed” job is equal to setting the status of the job to “processed”


- Once a job is processed, it will be ignored by the job processing service going forward, but is still able to be
requested by the GET routes defined above


- The polling interval for the job processing service should be configurable


## Task:


Create a technical design for the system described in the scenario above. You are free to choose whatever languages,
platforms, and other technologies you determine are appropriate. Please supply the following:


- An architectural diagram of your solution (made with Visio or a free alternative). Please label all components of
your system’s architecture


- A data flow diagram, detailing what your data models looks like and how data travels through your system from
end to end


- Any applicable database schemas if they exist


- Code samples showing how parts of your system work. These samples must run without any errors, but you do
not need to submit the code for the full working system. You can “fake” certain parts of the system by providing
dummy or test data.


## Notes:

- The architecture of your system is the most important part of this exercise. Please begin with the diagrams
described above and save the code samples for the end.


- You will have 2 days to complete this exercise. Within 48 hours of receiving this test, please submit your results
by sending us a GitHub link so that we can download everything.


- If anything in the question is unclear, make assumptions, write them down, and continue


- See below for some sample JSON objects representing jobs. Note that you can design your data objects however
you want, the below are just examples.