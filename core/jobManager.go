package core

type FileUpload struct {
  JobID int
  File []byte
}

func JobManager(fileUpload <-chan FileUpload, taskNotify chan<- int, dataStore Database, logger Logger){

  //create notification hub


  //listen for user sessions and hook them in
  go func(){
    for {
      select {
        case s := <-fileUpload:
        logger.Info("File Upload Received for task %v", s.JobID)

      }
    }
  }()

}
