// Converts the local YYYY-MM-DD date from the input field to a UTC ISO string for database storage.
    function dateToISO(dateString) {
        if (!dateString) return '';
        // Creates a Date object based on the user's local timezone
        const localDate = new Date(dateString); 
        // Converts it to a UTC string (the format you want to store in the DB)
        return localDate.toISOString(); 
    }

    // Converts the UTC ISO string (from DB) to the YYYY-MM-DD format required by <input type="date"> 
    // This displays the date correctly in the user's local timezone.
    function isoToDateInput(isoString) {
        if (!isoString) return '';
        const date = new Date(isoString); 
        const year = date.getFullYear();
        // Use +1 for getMonth() as it is zero-indexed
        const month = String(date.getMonth() + 1).padStart(2, '0'); 
        const day = String(date.getDate()).padStart(2, '0');
        return `${year}-${month}-${day}`;
    }