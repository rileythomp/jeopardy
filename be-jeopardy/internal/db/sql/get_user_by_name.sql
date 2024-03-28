select id, email,
COALESCE(raw_user_meta_data->>'display_name', raw_user_meta_data->>'full_name') as display_name,
COALESCE(raw_user_meta_data->>'user_img_url', raw_user_meta_data->>'avatar_url') as img_url,
COALESCE((raw_user_meta_data->>'profile_public')::boolean, true) as public,
created_at
from auth.users
where COALESCE(raw_user_meta_data->>'display_name', raw_user_meta_data->>'full_name') = $1
and COALESCE((raw_user_meta_data->>'profile_public')::boolean, true) = true;