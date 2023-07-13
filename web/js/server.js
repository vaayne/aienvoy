import PocketBase from "https://cdnjs.cloudflare.com/ajax/libs/pocketbase/0.15.2/pocketbase.es.mjs";

const client = new PocketBase();



class Services {
    static async Login(user, pass) {
        await client
            .collection("users")
            .authWithPassword(user, pass);
        if (pb.authStore.isValid) {
            document.cookie = `token=${pb.authStore.token}`;
            return true;
        }
        return false;
    }
}

